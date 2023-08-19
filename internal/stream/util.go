package stream

import (
	"context"
	"fmt"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/net"
	"github.com/alist-org/alist/v3/pkg/http_range"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
)

//func GetFileStreamFromLink(file model.Obj, link *model.Link, r http_range.Range) (*FileStream, error) {
//	readCloser, mimetype, err := GetReadCloserFromLink(file, link, r)
//	if err != nil {
//		return nil, err
//	}
//	// if can't get mimetype, use default application/octet-stream
//
//	stream := &FileStream{
//		Obj:      file,
//		Reader:   readCloser,
//		Mimetype: mimetype,
//		Closers:  *utils.NewClosers(readCloser),
//	}
//	return stream, nil
//}
//func GetReadCloserFromLink(ctx context.Context, file model.Obj, link *model.Link, r http_range.Range) (readCloser io.ReadCloser, mimetype string, err error) {
//	if r.Length == 0 {
//		r = http_range.Range{Length: -1}
//	}
//	mimetype = utils.GetMimeType(file.GetName())
//	if link.RangeReadCloser != nil {
//		readCloser, err = link.RangeReadCloser.RangeRead(ctx, r)
//		if err != nil {
//			return nil, "", err
//		}
//	} else if link.MFile != nil {
//		if r.Start == 0 && r.Length == -1 {
//			readCloser = link.MFile
//		} else {
//			//pay attention to multi-thread usage
//			_, err := link.MFile.Seek(r.Start, io.SeekStart)
//			if err != nil {
//				return nil, "", errs.NewErr(err, "GetReadCloserFromLink failed")
//			}
//			readCloser = link.MFile
//		}
//
//	} else if link.Concurrency != 0 || link.PartSize != 0 {
//		size := file.GetSize()
//		header := net.ProcessHeader(&http.Header{}, &link.Header)
//		down := net.NewDownloader(func(d *net.Downloader) {
//			d.Concurrency = link.Concurrency
//			d.PartSize = link.PartSize
//		})
//		req := &net.HttpRequestParams{
//			URL:       link.URL,
//			Range:     r,
//			Size:      size,
//			HeaderRef: header,
//		}
//		rc, err := down.Download(ctx, req)
//		if err != nil {
//			return nil, "", errs.NewErr(err, "GetReadCloserFromLink failed")
//		}
//		readCloser = *rc
//
//	} else {
//		req, err := http.NewRequest(http.MethodGet, link.URL, nil)
//		if err != nil {
//			return nil, "", errors.Wrapf(err, "failed to create request for %s", link.URL)
//		}
//		for h, val := range link.Header {
//			req.Header[h] = val
//		}
//		res, err := net.HttpClient().Do(req)
//		if err != nil {
//			return nil, "", errors.Wrapf(err, "failed to get response for %s", link.URL)
//		}
//		mt := res.Header.Get("Content-Type")
//		if mt != "" && strings.ToLower(mt) != "application/octet-stream" {
//			mimetype = mt
//		}
//		readCloser = res.Body
//	}
//	if mimetype == "" {
//		mimetype = "application/octet-stream"
//	}
//	return readCloser, mimetype, err
//}

func GetRangeReadCloserFromLink(size int64, link *model.Link) (model.RangeReadCloserIF, error) {
	if len(link.URL) == 0 {
		return nil, fmt.Errorf("can't create RangeReadCloser since URL is empty in link")
	}
	//remoteClosers := utils.EmptyClosers()
	rangeReaderFunc := func(ctx context.Context, r http_range.Range) (io.ReadCloser, error) {
		if link.Concurrency != 0 || link.PartSize != 0 {
			header := net.ProcessHeader(&http.Header{}, &link.Header)
			down := net.NewDownloader(func(d *net.Downloader) {
				d.Concurrency = link.Concurrency
				d.PartSize = link.PartSize
			})
			req := &net.HttpRequestParams{
				URL:       link.URL,
				Range:     r,
				Size:      size,
				HeaderRef: header,
			}
			rc, err := down.Download(ctx, req)
			if err != nil {
				return nil, errs.NewErr(err, "GetReadCloserFromLink failed")
			}
			return *rc, nil

		}
		if len(link.URL) > 0 {
			response, err := RequestRangedHttp(ctx, link, r.Start, r.Length)
			if err != nil {
				return nil, fmt.Errorf("http request failure,status: %d err:%s", response.StatusCode, err)
			}
			if r.Start == 0 && (r.Length == -1 || r.Length == size) || response.StatusCode == http.StatusPartialContent ||
				checkContentRange(&response.Header, size, r.Start) {
				return response.Body, nil
			} else if response.StatusCode == http.StatusOK {
				log.Warnf("remote http server not supporting range request, expect low perfromace!")
				readCloser, err := net.GetRangedHttpReader(response.Body, r.Start, r.Length)
				if err != nil {
					return nil, err
				}
				return readCloser, nil

			}

			return response.Body, nil
		}

		return nil, errs.NotSupport
	}
	resultRangeReadCloser := model.RangeReadCloser{RangeReader: rangeReaderFunc}
	return &resultRangeReadCloser, nil
}

func RequestRangedHttp(ctx context.Context, link *model.Link, offset, length int64) (*http.Response, error) {
	header := net.ProcessHeader(&http.Header{}, &link.Header)
	header = http_range.ApplyRangeToHttpHeader(http_range.Range{Start: offset, Length: length}, header)

	return net.RequestHttp(ctx, "GET", header, link.URL)
}

// 139 cloud does not properly return 206 http status code, add a hack here
func checkContentRange(header *http.Header, size, offset int64) bool {
	r, err2 := http_range.ParseRange(header.Get("Content-Range"), size)
	if err2 != nil {
		log.Warnf("exception trying to parse Content-Range, will ignore,err=%s", err2)
	}
	if len(r) == 1 && r[0].Start == offset {
		return true
	}
	return false
}

// LimitSeekReader returns a Reader that reads from rs
// but stops with EOF after n bytes.
// The underlying implementation is a *LimitedReader.
func LimitSeekReader(rs io.ReadSeeker, n int64) io.Reader { return &LimitedReadSeeker{rs, n} }

// A LimitedReadSeeker reads from Rs but limits the amount of
// data returned to just N bytes. Each call to Read
// updates N to reflect the new amount remaining.
// Read returns EOF when N <= 0 or when the underlying R returns EOF.
// it will also try to seek to 0 when got EOF or reached N
type LimitedReadSeeker struct {
	Rs io.ReadSeeker // underlying reader
	N  int64         // max bytes remaining
}

func (l *LimitedReadSeeker) Read(p []byte) (n int, err error) {
	if l.N <= 0 {
		err = l.trySeek0()
		if err != nil {
			return 0, err
		}
		return 0, io.EOF
	}
	if int64(len(p)) > l.N {
		p = p[0:l.N]
	}
	n, err = l.Rs.Read(p)
	l.N -= int64(n)
	return
}
func (l *LimitedReadSeeker) trySeek0() error {
	_, err := l.Rs.Seek(0, io.SeekStart)
	if err != nil {
		return errs.NewErr(err, "failed to seek to 0: %+v", err)
	}
	return nil
}
