package fs

import (
	"context"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/operations"
	"github.com/pkg/errors"
)

func Get(ctx context.Context, path string) (driver.FileInfo, error) {
	account, actualPath, err := operations.GetAccountAndActualPath(path)
	if err != nil {
		return nil, errors.WithMessage(err, "failed get account")
	}
	return account.File(ctx, actualPath)
}
