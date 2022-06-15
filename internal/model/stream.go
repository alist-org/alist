package model

import (
	"io"
)

type FileStream struct {
	FileInfo
	io.ReadCloser
	Mimetype string
}

func (f FileStream) GetMimetype() string {
	return f.Mimetype
}
