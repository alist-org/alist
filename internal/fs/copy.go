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

var copyTaskManager = task.NewTaskManager()

func CopyBetween2Accounts(ctx context.Context, srcAccount, dstAccount driver.Driver, srcPath, dstPath string, setStatus func(status string)) error {
	setStatus("getting src object")
	srcObj, err := operations.Get(ctx, srcAccount, srcPath)
	if err != nil {
		return errors.WithMessagef(err, "failed get src [%s] file", srcPath)
	}
	if srcObj.IsDir() {
		setStatus("src object is dir, listing files")
		files, err := operations.List(ctx, srcAccount, srcPath)
		if err != nil {
			return errors.WithMessagef(err, "failed list src [%s] files", srcPath)
		}
		for _, file := range files {
			if utils.IsCanceled(ctx) {
				return nil
			}
			srcFilePath := stdpath.Join(srcPath, file.GetName())
			dstFilePath := stdpath.Join(dstPath, file.GetName())
			copyTaskManager.Add(fmt.Sprintf("copy %s to %s", srcFilePath, dstFilePath), func(task *task.Task) error {
				return CopyBetween2Accounts(ctx, srcAccount, dstAccount, srcFilePath, dstFilePath, task.SetStatus)
			})
		}
	} else {
		copyTaskManager.Add(fmt.Sprintf("copy %s to %s", srcPath, dstPath), func(task *task.Task) error {
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
