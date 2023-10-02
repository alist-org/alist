package fs

import (
	"context"
	"fmt"
	"github.com/alist-org/alist/v3/internal/model"
	"sync/atomic"

	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/pkg/task"
	"github.com/pkg/errors"
)

var UploadTaskManager = task.NewTaskManager(3, func(tid *uint64) {
	atomic.AddUint64(tid, 1)
})

// putAsTask add as a put task and return immediately
func putAsTask(dstDirPath string, file model.FileStreamer) error {
	storage, dstDirActualPath, err := op.GetStorageAndActualPath(dstDirPath)
	if err != nil {
		return errors.WithMessage(err, "failed get storage")
	}
	if storage.Config().NoUpload {
		return errors.WithStack(errs.UploadNotSupported)
	}
	if file.NeedStore() {
		_, err := file.CacheFullInTempFile()
		if err != nil {
			return errors.Wrapf(err, "failed to create temp file")
		}
		//file.SetReader(tempFile)
		//file.SetTmpFile(tempFile)
	}
	UploadTaskManager.Submit(task.WithCancelCtx(&task.Task[uint64]{
		Name: fmt.Sprintf("upload %s to [%s](%s)", file.GetName(), storage.GetStorage().MountPath, dstDirActualPath),
		Func: func(task *task.Task[uint64]) error {
			return op.Put(task.Ctx, storage, dstDirActualPath, file, task.SetProgress, true)
		},
	}))
	return nil
}

// putDirect put the file and return after finish
func putDirectly(ctx context.Context, dstDirPath string, file model.FileStreamer, lazyCache ...bool) error {
	storage, dstDirActualPath, err := op.GetStorageAndActualPath(dstDirPath)
	if err != nil {
		return errors.WithMessage(err, "failed get storage")
	}
	if storage.Config().NoUpload {
		return errors.WithStack(errs.UploadNotSupported)
	}
	return op.Put(ctx, storage, dstDirActualPath, file, nil, lazyCache...)
}
