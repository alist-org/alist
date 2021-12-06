package model

import "io"

type FileStream struct {
	File       io.ReadCloser
	Size       uint64
	ParentPath string
	Name       string
	MIMEType   string
}

func (file FileStream) Read(p []byte) (n int, err error) {
	return file.File.Read(p)
}

func (file FileStream) GetMIMEType() string {
	return file.MIMEType
}

func (file FileStream) GetSize() uint64 {
	return file.Size
}

func (file FileStream) Close() error {
	return file.File.Close()
}

func (file FileStream) GetFileName() string {
	return file.Name
}

func (file FileStream) GetParentPath() string {
	return file.ParentPath
}