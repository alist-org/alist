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

var CopyTaskManager = task.NewTaskManager(3, func(tid *uint64) {
	atomic.AddUint64(tid, 1)
})

// Copy if in the same storage, call move method
// if not, add copy task
func _copy(ctx context.Context, srcObjPath, dstDirPath string) (bool, error) {
	srcStorage, srcObjActualPath, err := operations.GetStorageAndActualPath(srcObjPath)
	if err != nil {
		return false, errors.WithMessage(err, "failed get src storage")
	}
	dstStorage, dstDirActualPath, err := operations.GetStorageAndActualPath(dstDirPath)
	if err != nil {
		return false, errors.WithMessage(err, "failed get dst storage")
	}
	// copy if in the same storage, just call driver.Copy
	if srcStorage.GetStorage() == dstStorage.GetStorage() {
		return false, operations.Copy(ctx, srcStorage, srcObjActualPath, dstDirActualPath)
	}
	// not in the same storage
	CopyTaskManager.Submit(task.WithCancelCtx(&task.Task[uint64]{
		Name: fmt.Sprintf("copy [%s](%s) to [%s](%s)", srcStorage.GetStorage().VirtualPath, srcObjActualPath, dstStorage.GetStorage().VirtualPath, dstDirActualPath),
		Func: func(task *task.Task[uint64]) error {
			return copyBetween2Storages(task, srcStorage, dstStorage, srcObjActualPath, dstDirActualPath)
		},
	}))
	return true, nil
}

func copyBetween2Storages(t *task.Task[uint64], srcStorage, dstStorage driver.Driver, srcObjPath, dstDirPath string) error {
	t.SetStatus("getting src object")
	srcObj, err := operations.Get(t.Ctx, srcStorage, srcObjPath)
	if err != nil {
		return errors.WithMessagef(err, "failed get src [%s] file", srcObjPath)
	}
	if srcObj.IsDir() {
		t.SetStatus("src object is dir, listing objs")
		objs, err := operations.List(t.Ctx, srcStorage, srcObjPath)
		if err != nil {
			return errors.WithMessagef(err, "failed list src [%s] objs", srcObjPath)
		}
		for _, obj := range objs {
			if utils.IsCanceled(t.Ctx) {
				return nil
			}
			srcObjPath := stdpath.Join(srcObjPath, obj.GetName())
			dstObjPath := stdpath.Join(dstDirPath, obj.GetName())
			CopyTaskManager.Submit(task.WithCancelCtx(&task.Task[uint64]{
				Name: fmt.Sprintf("copy [%s](%s) to [%s](%s)", srcStorage.GetStorage().VirtualPath, srcObjPath, dstStorage.GetStorage().VirtualPath, dstObjPath),
				Func: func(t *task.Task[uint64]) error {
					return copyBetween2Storages(t, srcStorage, dstStorage, srcObjPath, dstObjPath)
				},
			}))
		}
	} else {
		CopyTaskManager.Submit(task.WithCancelCtx(&task.Task[uint64]{
			Name: fmt.Sprintf("copy [%s](%s) to [%s](%s)", srcStorage.GetStorage().VirtualPath, srcObjPath, dstStorage.GetStorage().VirtualPath, dstDirPath),
			Func: func(t *task.Task[uint64]) error {
				return copyFileBetween2Storages(t, srcStorage, dstStorage, srcObjPath, dstDirPath)
			},
		}))
	}
	return nil
}

func copyFileBetween2Storages(tsk *task.Task[uint64], srcStorage, dstStorage driver.Driver, srcFilePath, dstDirPath string) error {
	srcFile, err := operations.Get(tsk.Ctx, srcStorage, srcFilePath)
	if err != nil {
		return errors.WithMessagef(err, "failed get src [%s] file", srcFilePath)
	}
	link, _, err := operations.Link(tsk.Ctx, srcStorage, srcFilePath, model.LinkArgs{})
	if err != nil {
		return errors.WithMessagef(err, "failed get [%s] link", srcFilePath)
	}
	stream, err := getFileStreamFromLink(srcFile, link)
	if err != nil {
		return errors.WithMessagef(err, "failed get [%s] stream", srcFilePath)
	}
	return operations.Put(tsk.Ctx, dstStorage, dstDirPath, stream, tsk.SetProgress)
}
