package model

import (
	"io"
	"time"
)

type Obj interface {
	GetSize() uint64
	GetName() string
	ModTime() time.Time
	IsDir() bool
	GetID() string
}

type FileStreamer interface {
	io.ReadCloser
	Obj
	GetMimetype() string
}

type URL interface {
	URL() string
}

type Thumbnail interface {
	Thumbnail() string
}
