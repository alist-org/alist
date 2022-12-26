package model

import (
	"io"
)

type FileStream struct {
	Obj
	io.ReadCloser
	Mimetype     string
	WebPutAsTask bool
	Old          Obj
}

func (f *FileStream) GetMimetype() string {
	return f.Mimetype
}

func (f *FileStream) NeedStore() bool {
	return f.WebPutAsTask
}

func (f *FileStream) GetReadCloser() io.ReadCloser {
	return f.ReadCloser
}

func (f *FileStream) SetReadCloser(rc io.ReadCloser) {
	f.ReadCloser = rc
}

func (f *FileStream) GetOld() Obj {
	return f.Old
}
