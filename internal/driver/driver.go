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
	// Init If already initialized, drop first
	// need to unmarshal string to addition first
	Init(ctx context.Context, account model.Account) error
	Drop(ctx context.Context) error
	// GetAccount just get raw account
	GetAccount() model.Account
	GetAddition() Additional
}

type Other interface {
	Other(ctx context.Context, data interface{}) (interface{}, error)
}

type Reader interface {
	List(ctx context.Context, path string) ([]FileInfo, error)
	Link(ctx context.Context, path string, args LinkArgs) (*Link, error)
	//Get(ctx context.Context, path string) (FileInfo, error) // maybe not need
}

type Writer interface {
	MakeDir(ctx context.Context, path string) error
	Move(ctx context.Context, srcPath, dstPath string) error
	Rename(ctx context.Context, srcPath, dstName string) error
	Copy(ctx context.Context, srcPath, dstPath string) error
	Remove(ctx context.Context, path string) error
	Put(ctx context.Context, parentPath string, stream FileStream) error
}
