package operations

import (
	"context"
	"github.com/alist-org/alist/v3/internal/driver"
)

// In order to facilitate adding some other things before and after file operations

// List files in storage, not contains virtual file
// TODO: cache, and prevent cache breakdown
func List(ctx context.Context, account driver.Driver, path string) ([]driver.FileInfo, error) {
	return account.List(ctx, path)
}

func Get(ctx context.Context, account driver.Driver, path string) (driver.FileInfo, error) {
	return account.Get(ctx, path)
}

// Link get link, if is a url. show have an expiry time
func Link(ctx context.Context, account driver.Driver, path string, args driver.LinkArgs) (*driver.Link, error) {
	return account.Link(ctx, path, args)
}

func MakeDir(ctx context.Context, account driver.Driver, path string) error {
	return account.MakeDir(ctx, path)
}

func Move(ctx context.Context, account driver.Driver, srcPath, dstPath string) error {
	return account.Move(ctx, srcPath, dstPath)
}

func Rename(ctx context.Context, account driver.Driver, srcPath, dstName string) error {
	return account.Rename(ctx, srcPath, dstName)
}

// Copy Just copy file[s] in an account
func Copy(ctx context.Context, account driver.Driver, srcPath, dstPath string) error {
	return account.Copy(ctx, srcPath, dstPath)
}

func Remove(ctx context.Context, account driver.Driver, path string) error {
	return account.Remove(ctx, path)
}

func Put(ctx context.Context, account driver.Driver, parentPath string, file driver.FileStream) error {
	return account.Put(ctx, parentPath, file)
}
