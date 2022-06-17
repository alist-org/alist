package fs

import (
	"context"
	"fmt"
	"github.com/alist-org/alist/v3/pkg/task"
	"github.com/alist-org/alist/v3/pkg/utils"
	stdpath "path"

	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/operations"
	"github.com/pkg/errors"
)

var CopyTaskManager = task.NewTaskManager()

// Copy if in an account, call move method
// if not, add copy task
func Copy(ctx context.Context, account driver.Driver, srcPath, dstPath string) (bool, error) {
	srcAccount, srcActualPath, err := operations.GetAccountAndActualPath(srcPath)
	if err != nil {
		return false, errors.WithMessage(err, "failed get src account")
	}
	dstAccount, dstActualPath, err := operations.GetAccountAndActualPath(dstPath)
	if err != nil {
		return false, errors.WithMessage(err, "failed get dst account")
	}
	// copy if in an account, just call driver.Copy
	if srcAccount.GetAccount() == dstAccount.GetAccount() {
		return false, operations.Copy(ctx, account, srcActualPath, dstActualPath)
	}
	// not in an account
	CopyTaskManager.Add(fmt.Sprintf("copy [%s](%s) to [%s](%s)", srcAccount.GetAccount().VirtualPath, srcActualPath, dstAccount.GetAccount().VirtualPath, dstActualPath), func(task *task.Task) error {
		return CopyBetween2Accounts(task.Ctx, srcAccount, dstAccount, srcActualPath, dstActualPath, task.SetStatus)
	})
	return true, nil
}

func CopyBetween2Accounts(ctx context.Context, srcAccount, dstAccount driver.Driver, srcPath, dstPath string, setStatus func(status string)) error {
	setStatus("getting src object")
	srcObj, err := operations.Get(ctx, srcAccount, srcPath)
	if err != nil {
		return errors.WithMessagef(err, "failed get src [%s] file", srcPath)
	}
	if srcObj.IsDir() {
		setStatus("src object is dir, listing objs")
		objs, err := operations.List(ctx, srcAccount, srcPath)
		if err != nil {
			return errors.WithMessagef(err, "failed list src [%s] objs", srcPath)
		}
		for _, obj := range objs {
			if utils.IsCanceled(ctx) {
				return nil
			}
			srcObjPath := stdpath.Join(srcPath, obj.GetName())
			dstObjPath := stdpath.Join(dstPath, obj.GetName())
			CopyTaskManager.Add(fmt.Sprintf("copy [%s](%s) to [%s](%s)", srcAccount.GetAccount().VirtualPath, srcObjPath, dstAccount.GetAccount().VirtualPath, dstObjPath), func(task *task.Task) error {
				return CopyBetween2Accounts(ctx, srcAccount, dstAccount, srcObjPath, dstObjPath, task.SetStatus)
			})
		}
	} else {
		CopyTaskManager.Add(fmt.Sprintf("copy [%s](%s) to [%s](%s)", srcAccount.GetAccount().VirtualPath, srcPath, dstAccount.GetAccount().VirtualPath, dstPath), func(task *task.Task) error {
			return CopyFileBetween2Accounts(task.Ctx, srcAccount, dstAccount, srcPath, dstPath, func(percentage float64) {
				task.SetStatus(fmt.Sprintf("uploading: %2.f%", percentage))
			})
		})
	}
	return nil
}

func CopyFileBetween2Accounts(ctx context.Context, srcAccount, dstAccount driver.Driver, srcPath, dstPath string, up driver.UpdateProgress) error {
	srcFile, err := operations.Get(ctx, srcAccount, srcPath)
	if err != nil {
		return errors.WithMessagef(err, "failed get src [%s] file", srcPath)
	}
	link, err := operations.Link(ctx, srcAccount, srcPath, model.LinkArgs{})
	if err != nil {
		return errors.WithMessagef(err, "failed get [%s] link", srcPath)
	}
	stream, err := getFileStreamFromLink(srcFile, link)
	if err != nil {
		return errors.WithMessagef(err, "failed get [%s] stream", srcPath)
	}
	return operations.Put(ctx, dstAccount, dstPath, stream, up)
}
