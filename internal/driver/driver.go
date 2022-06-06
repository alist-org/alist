package driver

import (
	"context"
)

type Driver interface {
	Reader
	Writer
	Other
}

type Reader interface {
	File(ctx context.Context, path string) (*FileInfo, error)
	List(ctx context.Context, path string) ([]FileInfo, error)
	Link(ctx context.Context, args LinkArgs) (*Link, error)
}

type Writer interface {
	MakeDir(ctx context.Context, path string) error
	Move(ctx context.Context, src, dst string) error
	Rename(ctx context.Context, src, dst string) error
	Copy(ctx context.Context, src, dst string) error
	Remove(ctx context.Context, path string) error
	Put(ctx context.Context, stream FileStream, parent string) error
}

type Other interface {
	Init(ctx context.Context) error
	Update(ctx context.Context) error
	Drop(ctx context.Context) error
}
