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
	IP     string
	Header http.Header
	Type   string
}

type Link struct {
	URL        string         `json:"url"`
	Header     http.Header    `json:"header"` // needed header
	Data       io.ReadCloser  // return file reader directly
	Status     int            // status maybe 200 or 206, etc
	FilePath   *string        // local file, return the filepath
	Expiration *time.Duration // url expiration time
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
