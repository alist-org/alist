package fs

import (
	"context"
	"fmt"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/operations"
	"github.com/alist-org/alist/v3/pkg/task"
	"github.com/pkg/errors"
)

var UploadTaskManager = task.NewTaskManager()

// Put add as a put task
func Put(ctx context.Context, account driver.Driver, dstDir string, file model.FileStreamer) error {
	account, actualParentPath, err := operations.GetAccountAndActualPath(dstDir)
	if account.Config().NoUpload {
		return errors.WithStack(ErrUploadNotSupported)
	}
	if err != nil {
		return errors.WithMessage(err, "failed get account")
	}
	UploadTaskManager.Submit(fmt.Sprintf("upload %s to [%s](%s)", file.GetName(), account.GetAccount().VirtualPath, actualParentPath), func(task *task.Task) error {
		return operations.Put(task.Ctx, account, actualParentPath, file, nil)
	})
	return nil
}
