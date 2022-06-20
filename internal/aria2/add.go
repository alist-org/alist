package aria2

import (
	"context"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/fs"
	"github.com/alist-org/alist/v3/internal/operations"
	"github.com/pkg/errors"
)

func AddURI(ctx context.Context, uri []string, dstPath string, parentPath string) error {
	// check account
	account, actualParentPath, err := operations.GetAccountAndActualPath(parentPath)
	if err != nil {
		return errors.WithMessage(err, "failed get account")
	}
	// check is it could upload
	if account.Config().NoUpload {
		return errors.WithStack(fs.ErrUploadNotSupported)
	}
	// check path is valid
	obj, err := operations.Get(ctx, account, actualParentPath)
	if err != nil {
		if !errors.Is(errors.Cause(err), driver.ErrorObjectNotFound) {
			return errors.WithMessage(err, "failed get object")
		}
	} else {
		if !obj.IsDir() {
			// can't add to a file
			return errors.WithStack(fs.ErrNotFolder)
		}
	}
	// add aria2 task
	return nil
}
