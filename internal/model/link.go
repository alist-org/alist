package model

import (
	"io"
	"net/http"
	"time"
)

type LinkArgs struct {
	IP     string
	Header http.Header
}

type Link struct {
	URL        string         `json:"url"`
	Header     http.Header    `json:"header"` // needed header
	Data       io.ReadCloser  // return file reader directly
	Status     int            // status maybe 200 or 206, etc
	FilePath   *string        // local file, return the filepath
	Expiration *time.Duration // url expiration time
}
