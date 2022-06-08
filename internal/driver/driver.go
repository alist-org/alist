package driver

import (
	"context"
	"github.com/alist-org/alist/v3/internal/model"
)

type Driver interface {
	Meta
	Reader
	Writer
	Other
}

type Meta interface {
	Config() Config
	Init(ctx context.Context, account model.Account) error
	Update(ctx context.Context, account model.Account) error
	Drop(ctx context.Context) error
	// GetAccount transform additional field to string and assign to account's addition
	GetAccount() model.Account
	GetAddition() Additional
}

type Other interface {
	Other(ctx context.Context, data interface{}) (interface{}, error)
}

type Reader interface {
	File(ctx context.Context, path string) (FileInfo, error)
	List(ctx context.Context, path string) ([]FileInfo, error)
	Link(ctx context.Context, args LinkArgs) (*Link, error)
}

type Writer interface {
	MakeDir(ctx context.Context, path string) error
	Move(ctx context.Context, src, dst string) error
	Rename(ctx context.Context, src, dst string) error
	Copy(ctx context.Context, src, dst string) error
	Remove(ctx context.Context, path string) error
	Put(ctx context.Context, stream FileStream, parentPath string) error
}
