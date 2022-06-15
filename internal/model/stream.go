package model

import (
	"io"
)

type FileStream struct {
	Object
	io.ReadCloser
	Mimetype string
}

func (f FileStream) GetMimetype() string {
	return f.Mimetype
}
