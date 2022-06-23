package model

import (
	"io"
	"time"
)

type Obj interface {
	GetSize() int64
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

type SetID interface {
	SetID(id string)
}
