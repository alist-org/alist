package model

import (
	"io"
	"net/http"
	"time"

	"github.com/alist-org/alist/v3/pkg/http_range"
	"github.com/alist-org/alist/v3/pkg/utils"
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
	RangeReadCloser RangeReadCloser   `json:"-"`      // recommended way
	ReadSeekCloser  io.ReadSeekCloser `json:"-"`      // best for local,smb... file system, which exposes ReadSeekCloser

	Expiration *time.Duration // local cache expire Duration
	IPCacheKey bool           `json:"-"` // add ip to cache key
	//for accelerating request, use multi-thread downloading
	Concurrency int `json:"concurrency"`
	PartSize    int `json:"part_size"`
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
	Closers     *utils.Closers
}

type WriterFunc func(w io.Writer) error
type RangeReaderFunc func(httpRange http_range.Range) (io.ReadCloser, error)
