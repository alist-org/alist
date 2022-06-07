package driver

import (
	"io"
	"time"
)

type FileInfo interface {
	GetSize() uint64
	GetName() string
	ModTime() time.Time
	IsDir() bool
}

type FileStream interface {
	io.ReadCloser
	FileInfo
	GetMimetype() string
}

type URL interface {
	URL() string
}

type Thumbnail interface {
	Thumbnail() string
}
