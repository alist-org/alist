package model

import (
	"io"
)

type FileStream struct {
	Obj
	io.ReadCloser
	Mimetype string
}

func (f FileStream) GetMimetype() string {
	return f.Mimetype
}
