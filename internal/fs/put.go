package fs

import (
	"context"
	"fmt"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/operations"
	"github.com/alist-org/alist/v3/pkg/task"
	"github.com/pkg/errors"
	"sync/atomic"
)

var UploadTaskManager = task.NewTaskManager[uint64](3, func(tid *uint64) {
	atomic.AddUint64(tid, 1)
})

// putAsTask add as a put task and return immediately
func putAsTask(dstDirPath string, file model.FileStreamer) error {
	account, dstDirActualPath, err := operations.GetAccountAndActualPath(dstDirPath)
	if account.Config().NoUpload {
		return errors.WithStack(errs.UploadNotSupported)
	}
	if err != nil {
		return errors.WithMessage(err, "failed get account")
	}
	UploadTaskManager.Submit(task.WithCancelCtx(&task.Task[uint64]{
		Name: fmt.Sprintf("upload %s to [%s](%s)", file.GetName(), account.GetAccount().VirtualPath, dstDirActualPath),
		Func: func(task *task.Task[uint64]) error {
			return operations.Put(task.Ctx, account, dstDirActualPath, file, nil)
		},
	}))
	return nil
}

// putDirect put the file and return after finish
func putDirectly(ctx context.Context, dstDirPath string, file model.FileStreamer) error {
	account, dstDirActualPath, err := operations.GetAccountAndActualPath(dstDirPath)
	if account.Config().NoUpload {
		return errors.WithStack(errs.UploadNotSupported)
	}
	if err != nil {
		return errors.WithMessage(err, "failed get account")
	}
	return operations.Put(ctx, account, dstDirActualPath, file, nil)
}
