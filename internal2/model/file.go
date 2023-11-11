package model

import "io"

// File is basic file level accessing interface
type File interface {
	io.Reader
	io.ReaderAt
	io.Seeker
	io.Closer
}

type NopMFileIF interface {
	io.Reader
	io.ReaderAt
	io.Seeker
}
type NopMFile struct {
	NopMFileIF
}

func (NopMFile) Close() error { return nil }
func NewNopMFile(r NopMFileIF) File {
	return NopMFile{r}
}
