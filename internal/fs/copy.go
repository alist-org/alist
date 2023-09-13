package fs

import (
	"context"
	"fmt"
	"net/http"
	stdpath "path"
	"sync/atomic"

	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/internal/stream"
	"github.com/alist-org/alist/v3/pkg/task"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var CopyTaskManager = task.NewTaskManager(3, func(tid *uint64) {
	atomic.AddUint64(tid, 1)
})

// Copy if in the same storage, call move method
// if not, add copy task
func _copy(ctx context.Context, srcObjPath, dstDirPath string, lazyCache ...bool) (bool, error) {
	srcStorage, srcObjActualPath, err := op.GetStorageAndActualPath(srcObjPath)
	if err != nil {
		return false, errors.WithMessage(err, "failed get src storage")
	}
	dstStorage, dstDirActualPath, err := op.GetStorageAndActualPath(dstDirPath)
	if err != nil {
		return false, errors.WithMessage(err, "failed get dst storage")
	}
	// copy if in the same storage, just call driver.Copy
	if srcStorage.GetStorage() == dstStorage.GetStorage() {
		return false, op.Copy(ctx, srcStorage, srcObjActualPath, dstDirActualPath, lazyCache...)
	}
	if ctx.Value(conf.NoTaskKey) != nil {
		srcObj, err := op.Get(ctx, srcStorage, srcObjActualPath)
		if err != nil {
			return false, errors.WithMessagef(err, "failed get src [%s] file", srcObjPath)
		}
		if !srcObj.IsDir() {
			// copy file directly
			link, _, err := op.Link(ctx, srcStorage, srcObjActualPath, model.LinkArgs{
				Header: http.Header{},
			})
			if err != nil {
				return false, errors.WithMessagef(err, "failed get [%s] link", srcObjPath)
			}
			fs := stream.FileStream{
				Obj: srcObj,
				Ctx: ctx,
			}
			// any link provided is seekable
			ss, err := stream.NewSeekableStream(fs, link)
			if err != nil {
				return false, errors.WithMessagef(err, "failed get [%s] stream", srcObjPath)
			}
			return false, op.Put(ctx, dstStorage, dstDirActualPath, ss, nil, false)
		}
	}
	// not in the same storage
	CopyTaskManager.Submit(task.WithCancelCtx(&task.Task[uint64]{
		Name: fmt.Sprintf("copy [%s](%s) to [%s](%s)", srcStorage.GetStorage().MountPath, srcObjActualPath, dstStorage.GetStorage().MountPath, dstDirActualPath),
		Func: func(task *task.Task[uint64]) error {
			return copyBetween2Storages(task, srcStorage, dstStorage, srcObjActualPath, dstDirActualPath)
		},
	}))
	return true, nil
}

func copyBetween2Storages(t *task.Task[uint64], srcStorage, dstStorage driver.Driver, srcObjPath, dstDirPath string) error {
	t.SetStatus("getting src object")
	srcObj, err := op.Get(t.Ctx, srcStorage, srcObjPath)
	if err != nil {
		return errors.WithMessagef(err, "failed get src [%s] file", srcObjPath)
	}
	if srcObj.IsDir() {
		t.SetStatus("src object is dir, listing objs")
		objs, err := op.List(t.Ctx, srcStorage, srcObjPath, model.ListArgs{})
		if err != nil {
			return errors.WithMessagef(err, "failed list src [%s] objs", srcObjPath)
		}
		for _, obj := range objs {
			if utils.IsCanceled(t.Ctx) {
				return nil
			}
			srcObjPath := stdpath.Join(srcObjPath, obj.GetName())
			dstObjPath := stdpath.Join(dstDirPath, srcObj.GetName())
			CopyTaskManager.Submit(task.WithCancelCtx(&task.Task[uint64]{
				Name: fmt.Sprintf("copy [%s](%s) to [%s](%s)", srcStorage.GetStorage().MountPath, srcObjPath, dstStorage.GetStorage().MountPath, dstObjPath),
				Func: func(t *task.Task[uint64]) error {
					return copyBetween2Storages(t, srcStorage, dstStorage, srcObjPath, dstObjPath)
				},
			}))
		}
	} else {
		CopyTaskManager.Submit(task.WithCancelCtx(&task.Task[uint64]{
			Name: fmt.Sprintf("copy [%s](%s) to [%s](%s)", srcStorage.GetStorage().MountPath, srcObjPath, dstStorage.GetStorage().MountPath, dstDirPath),
			Func: func(t *task.Task[uint64]) error {
				err := copyFileBetween2Storages(t, srcStorage, dstStorage, srcObjPath, dstDirPath)
				log.Debugf("copy file between storages: %+v", err)
				return err
			},
		}))
	}
	return nil
}

func copyFileBetween2Storages(tsk *task.Task[uint64], srcStorage, dstStorage driver.Driver, srcFilePath, dstDirPath string) error {
	srcFile, err := op.Get(tsk.Ctx, srcStorage, srcFilePath)
	if err != nil {
		return errors.WithMessagef(err, "failed get src [%s] file", srcFilePath)
	}
	link, _, err := op.Link(tsk.Ctx, srcStorage, srcFilePath, model.LinkArgs{
		Header: http.Header{},
	})
	if err != nil {
		return errors.WithMessagef(err, "failed get [%s] link", srcFilePath)
	}
	fs := stream.FileStream{
		Obj: srcFile,
		Ctx: tsk.Ctx,
	}
	// any link provided is seekable
	ss, err := stream.NewSeekableStream(fs, link)
	if err != nil {
		return errors.WithMessagef(err, "failed get [%s] stream", srcFilePath)
	}
	return op.Put(tsk.Ctx, dstStorage, dstDirPath, ss, tsk.SetProgress, true)
}
