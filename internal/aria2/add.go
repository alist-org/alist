package aria2

import (
	"context"
	"github.com/alist-org/alist/v3/conf"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/fs"
	"github.com/alist-org/alist/v3/internal/operations"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"path/filepath"
)

func AddURI(ctx context.Context, uri string, dstPath string, parentPath string) error {
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
	// call aria2 rpc
	options := map[string]interface{}{
		"dir": filepath.Join(conf.Conf.TempDir, "aria2", uuid.NewString()),
	}
	gid, err := client.AddURI([]string{uri}, options)
	if err != nil {
		return errors.Wrapf(err, "failed to add uri %s", uri)
	}
	// TODO add to task manager
	return nil
}
