package model

import (
	"io"
	"time"
)

type Object interface {
	GetSize() uint64
	GetName() string
	ModTime() time.Time
	IsDir() bool
	GetID() string
}

type FileStreamer interface {
	io.ReadCloser
	Object
	GetMimetype() string
}

type URL interface {
	URL() string
}

type Thumbnail interface {
	Thumbnail() string
}
