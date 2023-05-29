package model

import (
	"io"
	"net/http"
	"time"
)

type ListArgs struct {
	ReqPath string
}

type LinkArgs struct {
	IP      string
	Header  http.Header
	Type    string
	HttpReq *http.Request
}

type Link struct {
	URL        string         `json:"url"`
	Header     http.Header    `json:"header"` // needed header (for url) or response header(for data or writer)
	Data       io.ReadCloser  // return file reader directly
	Status     int            // status maybe 200 or 206, etc
	FilePath   *string        // local file, return the filepath
	Expiration *time.Duration // url expiration time
	//Handle     func(w http.ResponseWriter, r *http.Request) error `json:"-"` // custom handler
	Writer WriterFunc `json:"-"` // custom writer
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

type WriterFunc func(w io.Writer) error
