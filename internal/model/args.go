package model

import (
	"github.com/alist-org/alist/v3/pkg/http_range"
	"io"
	"net/http"
	"time"
)

type ListArgs struct {
	ReqPath           string
	S3ShowPlaceholder bool
}

type LinkArgs struct {
	IP      string
	Header  http.Header
	Type    string
	HttpReq *http.Request
}

type Link struct {
	URL             string            `json:"url"`
	Header          http.Header       `json:"header"` // needed header (for url) or response header(for data or writer)
	RangeReadCloser RangeReadCloser   // recommended way
	ReadSeekCloser  io.ReadSeekCloser // best for local,smb.. file system, which exposes ReadSeekCloser

	Expiration *time.Duration // local cache expire Duration
	IPCacheKey bool           // add ip to cache key
	//for accelerating request, use multi-thread downloading
	Concurrency int
	PartSize    int
}

type OtherArgs struct {
	Obj    Obj
	Method string
	Data   interface{}
}

type FsOtherArgs struct {
	Path   string      `json:"path" form:"path"`
	Method string      `json:"method" form:"method"`
	Data   interface{} `json:"data" form:"data"`
}
type RangeReadCloser struct {
	RangeReader RangeReaderFunc
	Closer      io.Closer
}

type Closers struct {
	closers []io.Closer
}

func (c *Closers) Close() (err error) {
	for _, closer := range c.closers {
		if closer != nil {
			_ = closer.Close()
		}
	}
	return nil
}
func (c *Closers) Add(closer io.Closer) {
	c.closers = append(c.closers, closer)
}

type WriterFunc func(w io.Writer) error
type RangeReaderFunc func(httpRange http_range.Range) (io.ReadCloser, error)
