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
	List(ctx context.Context, dir model.Object) ([]model.Object, error)
	Link(ctx context.Context, file model.Object, args model.LinkArgs) (*model.Link, error)
	//Get(ctx context.Context, path string) (FileInfo, error) // maybe not need
}

type Writer interface {
	// MakeDir make a folder named `dirName` in `parentDir`
	MakeDir(ctx context.Context, parentDir model.Object, dirName string) error
	// Move move `srcObject` to `dstDir`
	Move(ctx context.Context, srcObject, dstDir model.Object) error
	// Rename rename `srcObject` to `newName`
	Rename(ctx context.Context, srcObject model.Object, newName string) error
	// Copy copy `srcObject` to `dstDir`
	Copy(ctx context.Context, srcObject, dstDir model.Object) error
	// Remove remove `object`
	Remove(ctx context.Context, object model.Object) error
	// Put put `stream` to `parentDir`
	Put(ctx context.Context, parentDir model.Object, stream model.FileStreamer) error
}
