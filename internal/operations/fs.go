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
