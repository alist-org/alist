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
	URL    string        `json:"url"`
	Header http.Header   `json:"header"` // needed header (for url) or response header(for data or writer)
	Data   io.ReadCloser // will remove later

	RangeReadCloser RangeReadCloser   // recommended way
	ReadSeekCloser  io.ReadSeekCloser // best for local,smb.. file system, which exposes ReadSeekCloser

	Status     int            // TODO: remove
	Expiration *time.Duration // url expiration time
	IPCacheKey bool           // add ip to cache key
	//Handle     func(w http.ResponseWriter, r *http.Request) error `json:"-"` // custom handler
	Writer WriterFunc `json:"-"` // TODO: remove
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

type WriterFunc func(w io.Writer) error
type RangeReaderFunc func(httpRange http_range.Range) (io.ReadCloser, error)
