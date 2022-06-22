package fs

import (
	"context"
	"fmt"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/operations"
	"github.com/alist-org/alist/v3/pkg/task"
	"github.com/pkg/errors"
	"sync/atomic"
)

var UploadTaskManager = task.NewTaskManager[uint64, struct{}](3, func(tid *uint64) {
	atomic.AddUint64(tid, 1)
})

// Put add as a put task
func Put(ctx context.Context, account driver.Driver, dstDirPath string, file model.FileStreamer) error {
	account, dstDirActualPath, err := operations.GetAccountAndActualPath(dstDirPath)
	if account.Config().NoUpload {
		return errors.WithStack(ErrUploadNotSupported)
	}
	if err != nil {
		return errors.WithMessage(err, "failed get account")
	}
	UploadTaskManager.Submit(task.WithCancelCtx(&task.Task[uint64, struct{}]{
		Name: fmt.Sprintf("upload %s to [%s](%s)", file.GetName(), account.GetAccount().VirtualPath, dstDirActualPath),
		Func: func(task *task.Task[uint64, struct{}]) error {
			return operations.Put(task.Ctx, account, dstDirActualPath, file, nil)
		},
	}))
	return nil
}
