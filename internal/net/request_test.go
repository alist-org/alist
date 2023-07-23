package net

//no http range
//

import (
	"bytes"
	"context"
	"fmt"
	"github.com/alist-org/alist/v3/pkg/http_range"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"golang.org/x/exp/slices"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"testing"
)

var buf22MB = make([]byte, 1024*1024*22)

func dummyHttpRequest(data []byte, p http_range.Range) io.ReadCloser {

	end := p.Start + p.Length - 1

	if end >= int64(len(data)) {
		end = int64(len(data))
	}

	bodyBytes := data[p.Start:end]
	return io.NopCloser(bytes.NewReader(bodyBytes))
}

func TestDownloadOrder(t *testing.T) {
	buff := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	downloader, invocations, ranges := newDownloadRangeClient(buff)
	var con = 3
	var partSize int64 = 3
	d := NewDownloader(func(d *Downloader) {
		d.Concurrency = con
		d.PartSize = partSize
		d.HttpClient = downloader.HttpRequest
	})

	w := manager.NewWriteAtBuffer(make([]byte, len(buff)))
	var start, length int64 = 2, 10
	legnth2 := length
	if legnth2 == -1 {
		legnth2 = int64(len(buff)) - start
	}
	req := &HttpRequestParams{
		Range: http_range.Range{Start: start, Length: length},
		Size:  int64(len(buff)),
	}
	n, err := d.Download(context.Background(), w, req)

	if err != nil {
		t.Fatalf("expect no error, got %v", err)
	}
	if e, a := legnth2, n; e != a {
		t.Errorf("expect %d buffer length, got %d", e, a)
	}
	//int(legnth2/partSize)
	if e, a := 4, *invocations; e != a {
		t.Errorf("expect %v API calls, got %v", e, a)
	}

	expectRngs := []string{"2-3", "5-3", "8-3", "11-1"}
	for _, rng := range expectRngs {
		if !slices.Contains(*ranges, rng) {
			t.Errorf("expect range %v, but absent in return", rng)
		}
	}
	if e, a := expectRngs, *ranges; len(e) != len(a) {
		t.Errorf("expect %v ranges, got %v", e, a)
	}
}

type downloadCaptureClient struct {
	mockedHttpRequest    func(params *HttpRequestParams) (*http.Response, error)
	GetObjectInvocations int

	RetrievedRanges []string

	lock sync.Mutex
}

func (c *downloadCaptureClient) HttpRequest(params *HttpRequestParams) (*http.Response, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.GetObjectInvocations++

	if &params.Range != nil {
		c.RetrievedRanges = append(c.RetrievedRanges, fmt.Sprintf("%d-%d", params.Range.Start, params.Range.Length))
	}

	return c.mockedHttpRequest(params)
}

func newDownloadRangeClient(data []byte) (*downloadCaptureClient, *int, *[]string) {
	capture := &downloadCaptureClient{}

	capture.mockedHttpRequest = func(params *HttpRequestParams) (*http.Response, error) {
		start, fin := params.Range.Start, params.Range.Start+params.Range.Length
		if params.Range.Length == -1 || fin >= int64(len(data)) {
			fin = int64(len(data))
		}
		bodyBytes := data[start:fin]

		header := &http.Header{}
		header.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, fin-1, len(data)))
		return &http.Response{
			Body:          ioutil.NopCloser(bytes.NewReader(bodyBytes)),
			Header:        *header,
			ContentLength: int64(len(bodyBytes)),
		}, nil
	}

	return capture, &capture.GetObjectInvocations, &capture.RetrievedRanges
}
