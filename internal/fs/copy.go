package fs

import (
	"context"
	"fmt"
	stdpath "path"
	"sync/atomic"

	"github.com/alist-org/alist/v3/pkg/task"
	"github.com/alist-org/alist/v3/pkg/utils"

	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/operations"
	"github.com/pkg/errors"
)

var CopyTaskManager = task.NewTaskManager[uint64, struct{}](3, func(tid *uint64) {
	atomic.AddUint64(tid, 1)
})

// Copy if in an account, call move method
// if not, add copy task
func Copy(ctx context.Context, account driver.Driver, srcObjPath, dstDirPath string) (bool, error) {
	srcAccount, srcObjActualPath, err := operations.GetAccountAndActualPath(srcObjPath)
	if err != nil {
		return false, errors.WithMessage(err, "failed get src account")
	}
	dstAccount, dstDirActualPath, err := operations.GetAccountAndActualPath(dstDirPath)
	if err != nil {
		return false, errors.WithMessage(err, "failed get dst account")
	}
	// copy if in an account, just call driver.Copy
	if srcAccount.GetAccount() == dstAccount.GetAccount() {
		return false, operations.Copy(ctx, account, srcObjActualPath, dstDirActualPath)
	}
	// not in an account
	CopyTaskManager.Submit(task.WithCancelCtx(&task.Task[uint64, struct{}]{
		Name: fmt.Sprintf("copy [%s](%s) to [%s](%s)", srcAccount.GetAccount().VirtualPath, srcObjActualPath, dstAccount.GetAccount().VirtualPath, dstDirActualPath),
		Func: func(task *task.Task[uint64, struct{}]) error {
			return CopyBetween2Accounts(task, srcAccount, dstAccount, srcObjActualPath, dstDirActualPath)
		},
	}))
	return true, nil
}

func CopyBetween2Accounts(t *task.Task[uint64, struct{}], srcAccount, dstAccount driver.Driver, srcObjPath, dstDirPath string) error {
	t.SetStatus("getting src object")
	srcObj, err := operations.Get(t.Ctx, srcAccount, srcObjPath)
	if err != nil {
		return errors.WithMessagef(err, "failed get src [%s] file", srcObjPath)
	}
	if srcObj.IsDir() {
		t.SetStatus("src object is dir, listing objs")
		objs, err := operations.List(t.Ctx, srcAccount, srcObjPath)
		if err != nil {
			return errors.WithMessagef(err, "failed list src [%s] objs", srcObjPath)
		}
		for _, obj := range objs {
			if utils.IsCanceled(t.Ctx) {
				return nil
			}
			srcObjPath := stdpath.Join(srcObjPath, obj.GetName())
			dstObjPath := stdpath.Join(dstDirPath, obj.GetName())
			CopyTaskManager.Submit(task.WithCancelCtx(&task.Task[uint64, struct{}]{
				Name: fmt.Sprintf("copy [%s](%s) to [%s](%s)", srcAccount.GetAccount().VirtualPath, srcObjPath, dstAccount.GetAccount().VirtualPath, dstObjPath),
				Func: func(t *task.Task[uint64, struct{}]) error {
					return CopyBetween2Accounts(t, srcAccount, dstAccount, srcObjPath, dstObjPath)
				},
			}))
		}
	} else {
		CopyTaskManager.Submit(task.WithCancelCtx(&task.Task[uint64, struct{}]{
			Name: fmt.Sprintf("copy [%s](%s) to [%s](%s)", srcAccount.GetAccount().VirtualPath, srcObjPath, dstAccount.GetAccount().VirtualPath, dstDirPath),
			Func: func(t *task.Task[uint64, struct{}]) error {
				return CopyFileBetween2Accounts(t, srcAccount, dstAccount, srcObjPath, dstDirPath)
			},
		}))
	}
	return nil
}

func CopyFileBetween2Accounts(tsk *task.Task[uint64, struct{}], srcAccount, dstAccount driver.Driver, srcFilePath, dstDirPath string) error {
	srcFile, err := operations.Get(tsk.Ctx, srcAccount, srcFilePath)
	if err != nil {
		return errors.WithMessagef(err, "failed get src [%s] file", srcFilePath)
	}
	link, err := operations.Link(tsk.Ctx, srcAccount, srcFilePath, model.LinkArgs{})
	if err != nil {
		return errors.WithMessagef(err, "failed get [%s] link", srcFilePath)
	}
	stream, err := getFileStreamFromLink(srcFile, link)
	if err != nil {
		return errors.WithMessagef(err, "failed get [%s] stream", srcFilePath)
	}
	return operations.Put(tsk.Ctx, dstAccount, dstDirPath, stream, tsk.SetProgress)
}
