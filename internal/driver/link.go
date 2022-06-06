package driver

import (
	"io"
	"net/http"
)

type LinkArgs struct {
	Path   string
	IP     string
	Header http.Header
}

type Link struct {
	URL      string
	Header   http.Header
	Data     io.ReadCloser
	Status   int
	FilePath string
}
