package net

import (
	"context"
	"fmt"
	"github.com/alist-org/alist/v3/pkg/http_range"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

//inspired by "github.com/aws/aws-sdk-go-v2/service/s3/manager"

// DefaultDownloadPartSize is the default range of bytes to get at a time when
// using Download().
const DefaultDownloadPartSize = 1024 * 1024 * 10

// DefaultDownloadConcurrency is the default number of goroutines to spin up
// when using Download().
const DefaultDownloadConcurrency = 2

// DefaultPartBodyMaxRetries is the default number of retries to make when a part fails to download.
const DefaultPartBodyMaxRetries = 3

// flush data out when bufferSize of data is received.
const bufferSize = 64 * 1024

type Downloader struct {
	PartSize int64

	// PartBodyMaxRetries is the number of retry attempts to make for failed part downloads.
	PartBodyMaxRetries int

	// The number of goroutines to spin up in parallel when sending parts.
	// If this is set to zero, the DefaultDownloadConcurrency value will be used.
	//
	// Concurrency of 1 will download the parts sequentially.
	Concurrency int

	//RequestParam        HttpRequestParams
	HttpClient HttpRequestFunc
}
type HttpRequestFunc func(params *HttpRequestParams) (*http.Response, error)

func NewDownloader(options ...func(*Downloader)) *Downloader {
	d := &Downloader{
		HttpClient:         DefaultHttpRequestFunc,
		PartSize:           DefaultDownloadPartSize,
		PartBodyMaxRetries: DefaultPartBodyMaxRetries,
		Concurrency:        DefaultDownloadConcurrency,
	}
	for _, option := range options {
		option(d)
	}
	return d
}

// Download The Downloader makes multi-thread http requests to remote URL, each chunk(except last one) has PartSize,
// cache some data, then return Reader with assembled data
// Supports range, do not support unknown FileSize, and will fail if FileSize is incorrect
// memory usage is at about Concurrency*PartSize, use this wisely
func (d Downloader) Download(ctx context.Context, w io.WriterAt, p *HttpRequestParams) (n int64, err error) {

	var finalP HttpRequestParams
	awsutil.Copy(&finalP, p)
	if finalP.Range.Length == -1 {
		finalP.Range.Length = finalP.Size - finalP.Range.Start
	}
	impl := downloader{w: w, params: &finalP, cfg: d, ctx: ctx, pos: finalP.Range.Start}

	// Ensures we don't need nil checks later on

	impl.partBodyMaxRetries = d.PartBodyMaxRetries

	impl.totalBytes = -1
	if impl.cfg.Concurrency == 0 {
		impl.cfg.Concurrency = DefaultDownloadConcurrency
	}

	if impl.cfg.PartSize == 0 {
		impl.cfg.PartSize = DefaultDownloadPartSize
	}

	return impl.download()
}

/*// Download The Downloader makes multi-thread http requests to remote URL, and returns ReadCloser
// currently do not support unknown FileSize, and will fail if FileSize is incorrect
func (d Downloader) download1(r http_range.Range, header *http.Header) (io.ReadCloser, error) {
	err := d.check()
	if err != nil {
		return nil, err
	}
	var chunks, length int64
	if r.Length == -1 {
		length = d.FileSize - r.Start
	} else {
		length = r.Length
	}
	chunks = length/d.PartSize + 1
	if length%d.PartSize == 0 {
		chunks--
	}
	guard := make(chan struct{}, d.Concurrency)

	for i := 0; i < chunks; i++ {
		guard <- struct{}{} // would block if guard channel is already filled
		go func(n int) {
			worker(n)
			<-guard
		}(i)
	}
	chunks := 1

	utils.Retry(d.PartBodyMaxRetries, 200*time.Millisecond, d.HttpClient())

}*/

// downloader is the implementation structure used internally by Downloader.
type downloader struct {
	ctx context.Context
	cfg Downloader

	params *HttpRequestParams
	w      io.WriterAt

	wg sync.WaitGroup
	m  sync.Mutex

	pos        int64
	totalBytes int64
	written    int64
	err        error

	partBodyMaxRetries int
}

// download performs the implementation of the object download across ranged GETs.
func (d *downloader) download() (n int64, err error) {

	// Spin off first worker to check if provided size is correct
	d.getChunk()

	if total := d.getTotalBytes(); total >= 0 {
		if total != d.params.Size {
			d.err = fmt.Errorf("expect file size=%d unmatch remote report size=%d, need refresh cache", d.params.Size, total)
		}

		// Spin up workers
		ch := make(chan chunk, d.cfg.Concurrency)

		for i := 0; i < d.cfg.Concurrency; i++ {
			d.wg.Add(1)
			go d.downloadPart(ch)
		}

		// Assign work
		for d.getErr() == nil {
			end := d.params.Range.Start + d.params.Range.Length
			if d.pos >= end {
				break // We're finished queuing chunks
			}
			s := d.cfg.PartSize
			if s+d.pos > end {
				s = end - d.pos
			}
			// Queue the next range of bytes to read.
			ch <- chunk{w: d.w, start: d.pos, size: s, boundary: d.params.Range}
			d.pos += d.cfg.PartSize
		}

		// Wait for completion
		close(ch)
		d.wg.Wait()
	} else {
		//we did not get file size from http response
		d.err = fmt.Errorf("can't get total file size from remote HTTP Server")
	}
	// Return error
	return d.written, d.err
}

// downloadPart is an individual goroutine worker reading from the ch channel
// and performing a GetObject request on the data with a given byte range.
//
// If this is the first worker, this operation also resolves the total number
// of bytes to be read so that the worker manager knows when it is finished.
func (d *downloader) downloadPart(ch chan chunk) {
	defer d.wg.Done()
	for {
		chunk, ok := <-ch
		if !ok {
			break
		}
		if d.getErr() != nil {
			// Drain the channel if there is an error, to prevent deadlocking
			// of download producer.
			continue
		}

		if err := d.downloadChunk(chunk); err != nil {
			d.setErr(err)
		}
	}
}

// getChunk grabs a chunk of data from the body.
// Not thread safe. Should only be used when grabbing data on a single thread.
func (d *downloader) getChunk() {
	if d.getErr() != nil {
		return
	}

	finalSize := d.params.Range.Length
	//only request as big as PartSize
	if finalSize > d.cfg.PartSize {
		finalSize = d.cfg.PartSize
	}
	chunk := chunk{w: d.w, start: d.pos, size: finalSize, boundary: d.params.Range}
	d.pos += d.cfg.PartSize

	if err := d.downloadChunk(chunk); err != nil {
		d.setErr(err)
	}
}

// downloadChunk downloads the chunk from s3
func (d *downloader) downloadChunk(chunk chunk) error {
	var params HttpRequestParams
	awsutil.Copy(&params, d.params)

	// Get the next byte range of data
	params.Range = http_range.Range{Start: chunk.start, Length: chunk.size}

	var n int64
	var err error
	for retry := 0; retry <= d.partBodyMaxRetries; retry++ {
		n, err = d.tryDownloadChunk(&params, &chunk)
		if err == nil {
			break
		}
		// Check if the returned error is an errReadingBody.
		// If err is errReadingBody this indicates that an error
		// occurred while copying the http response body.
		// If this occurs we unwrap the err to set the underlying error
		// and attempt any remaining retries.
		if bodyErr, ok := err.(*errReadingBody); ok {
			err = bodyErr.Unwrap()
		} else {
			return err
		}

		chunk.cur = 0

		log.Debugf("object part body download interrupted %s, err, %v, retrying attempt %d",
			params.URL, err, retry)
	}

	d.incrWritten(n)

	return err
}

func (d *downloader) tryDownloadChunk(params *HttpRequestParams, w io.Writer) (int64, error) {

	resp, err := d.cfg.HttpClient(params)
	if err != nil {
		return 0, err
	}
	d.setTotalBytes(resp) // Set total if not yet set.

	var src io.Reader = resp.Body
	n, err := io.Copy(w, src)
	resp.Body.Close()
	if err != nil {
		return n, &errReadingBody{err: err}
	}

	return n, nil
}

// getTotalBytes is a thread-safe getter for retrieving the total byte status.
func (d *downloader) getTotalBytes() int64 {
	d.m.Lock()
	defer d.m.Unlock()

	return d.totalBytes
}

// setTotalBytes is a thread-safe setter for setting the total byte status.
// Will extract the object's total bytes from the Content-Range if the file
// will be chunked, or Content-Length. Content-Length is used when the response
// does not include a Content-Range. Meaning the object was not chunked. This
// occurs when the full file fits within the PartSize directive.
func (d *downloader) setTotalBytes(resp *http.Response) {
	d.m.Lock()
	defer d.m.Unlock()

	if d.totalBytes >= 0 {
		return
	}

	//		link.Header.Set("Content-Range", parseRange[0].ContentRange(file.GetSize()))
	//		link.Header.Set("Content-Length", strconv.FormatInt(parseRange[0].Length, 10))
	contentRange := resp.Header.Get("Content-Range")
	if len(contentRange) == 0 {
		// ContentRange is nil when the full file contents is provided, and
		// is not chunked. Use ContentLength instead.
		if resp.ContentLength > 0 {
			d.totalBytes = resp.ContentLength
			return
		}
	} else {
		parts := strings.Split(contentRange, "/")

		total := int64(-1)
		var err error
		// Checking for whether or not a numbered total exists
		// If one does not exist, we will assume the total to be -1, undefined,
		// and sequentially download each chunk until hitting a 416 error
		totalStr := parts[len(parts)-1]
		if totalStr != "*" {
			total, err = strconv.ParseInt(totalStr, 10, 64)
			if err != nil {
				d.err = err
				return
			}
		}

		d.totalBytes = total
	}
}

func (d *downloader) incrWritten(n int64) {
	d.m.Lock()
	defer d.m.Unlock()

	d.written += n
}

// getErr is a thread-safe getter for the error object
func (d *downloader) getErr() error {
	d.m.Lock()
	defer d.m.Unlock()

	return d.err
}

// setErr is a thread-safe setter for the error object
func (d *downloader) setErr(e error) {
	d.m.Lock()
	defer d.m.Unlock()

	d.err = e
}

// Chunk represents a single chunk of data to write by the worker routine.
// This structure also implements an io.SectionReader style interface for
// io.WriterAt, effectively making it an io.SectionWriter (which does not
// exist).
type chunk struct {
	w     io.WriterAt
	start int64
	size  int64
	cur   int64

	// Downloader takes range (start,length), but this chunk is requesting equal/sub range of it.
	// To convert the writer to reader eventually, we need to write within the boundary
	boundary http_range.Range
}

// Write wraps io.WriterAt for the chunk, writing from the chunk's start
// position to its end (or EOF).
//
// If a range is specified on the chunk the size will be ignored when writing.
// as the total size may not of be known ahead of time.
func (c *chunk) Write(p []byte) (n int, err error) {
	if c.start >= c.boundary.Start+c.boundary.Length {
		return 0, io.EOF
	}

	n, err = c.w.WriteAt(p, c.start+c.cur-c.boundary.Start)
	c.cur += int64(n)

	return
}

func DefaultHttpRequestFunc(params *HttpRequestParams) (*http.Response, error) {
	header := http_range.ApplyRangeToHttpHeader(params.Range, params.HeaderRef)

	res, err := RequestHttp("GET", header, params.URL)
	if err != nil {
		return nil, err
	}
	return res, nil
}

type HttpRequestParams struct {
	URL string
	//only want data within this range
	Range     http_range.Range
	HeaderRef *http.Header
	//total file size
	Size int64
}
type errReadingBody struct {
	err error
}

func (e *errReadingBody) Error() string {
	return fmt.Sprintf("failed to read part body: %v", e.err)
}

func (e *errReadingBody) Unwrap() error {
	return e.err
}
