package driver

import (
	"io"
	"time"
)

type FileInfo interface {
	GetName() string
	GetModTime() time.Time
	GetSize() int64
}

type FileStream interface {
	io.ReadCloser
	FileInfo
	GetMimetype() string
}
