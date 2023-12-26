package fs

import (
	"context"
	"fmt"
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/internal/stream"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/pkg/errors"
	"github.com/xhofe/tache"
	"net/http"
	stdpath "path"
)

type CopyTask struct {
	tache.Base
	Status                 string `json:"status"`
	srcStorage, dstStorage driver.Driver
	srcObjPath, dstDirPath string
}

func (t *CopyTask) GetName() string {
	return fmt.Sprintf("copy [%s](%s) to [%s](%s)",
		t.srcStorage.GetStorage().MountPath, t.srcObjPath, t.dstStorage.GetStorage().MountPath, t.dstDirPath)
}

func (t *CopyTask) GetStatus() string {
	return t.Status
}

func (t *CopyTask) Run() error {
	return copyBetween2Storages(t, t.srcStorage, t.dstStorage, t.srcObjPath, t.dstDirPath)
}

var CopyTaskManager *tache.Manager[*CopyTask]

// Copy if in the same storage, call move method
// if not, add copy task
func _copy(ctx context.Context, srcObjPath, dstDirPath string, lazyCache ...bool) (tache.TaskWithInfo, error) {
	srcStorage, srcObjActualPath, err := op.GetStorageAndActualPath(srcObjPath)
	if err != nil {
		return nil, errors.WithMessage(err, "failed get src storage")
	}
	dstStorage, dstDirActualPath, err := op.GetStorageAndActualPath(dstDirPath)
	if err != nil {
		return nil, errors.WithMessage(err, "failed get dst storage")
	}
	// copy if in the same storage, just call driver.Copy
	if srcStorage.GetStorage() == dstStorage.GetStorage() {
		return nil, op.Copy(ctx, srcStorage, srcObjActualPath, dstDirActualPath, lazyCache...)
	}
	if ctx.Value(conf.NoTaskKey) != nil {
		srcObj, err := op.Get(ctx, srcStorage, srcObjActualPath)
		if err != nil {
			return nil, errors.WithMessagef(err, "failed get src [%s] file", srcObjPath)
		}
		if !srcObj.IsDir() {
			// copy file directly
			link, _, err := op.Link(ctx, srcStorage, srcObjActualPath, model.LinkArgs{
				Header: http.Header{},
			})
			if err != nil {
				return nil, errors.WithMessagef(err, "failed get [%s] link", srcObjPath)
			}
			fs := stream.FileStream{
				Obj: srcObj,
				Ctx: ctx,
			}
			// any link provided is seekable
			ss, err := stream.NewSeekableStream(fs, link)
			if err != nil {
				return nil, errors.WithMessagef(err, "failed get [%s] stream", srcObjPath)
			}
			return nil, op.Put(ctx, dstStorage, dstDirActualPath, ss, nil, false)
		}
	}
	// not in the same storage
	t := &CopyTask{
		srcStorage: srcStorage,
		dstStorage: dstStorage,
		srcObjPath: srcObjActualPath,
		dstDirPath: dstDirActualPath,
	}
	CopyTaskManager.Add(t)
	return t, nil
}

func copyBetween2Storages(t *CopyTask, srcStorage, dstStorage driver.Driver, srcObjPath, dstDirPath string) error {
	t.Status = "getting src object"
	srcObj, err := op.Get(t.Ctx(), srcStorage, srcObjPath)
	if err != nil {
		return errors.WithMessagef(err, "failed get src [%s] file", srcObjPath)
	}
	if srcObj.IsDir() {
		t.Status = "src object is dir, listing objs"
		objs, err := op.List(t.Ctx(), srcStorage, srcObjPath, model.ListArgs{})
		if err != nil {
			return errors.WithMessagef(err, "failed list src [%s] objs", srcObjPath)
		}
		for _, obj := range objs {
			if utils.IsCanceled(t.Ctx()) {
				return nil
			}
			srcObjPath := stdpath.Join(srcObjPath, obj.GetName())
			dstObjPath := stdpath.Join(dstDirPath, srcObj.GetName())
			CopyTaskManager.Add(&CopyTask{
				srcStorage: srcStorage,
				dstStorage: dstStorage,
				srcObjPath: srcObjPath,
				dstDirPath: dstObjPath,
			})
		}
		t.Status = "src object is dir, added all copy tasks of objs"
		return nil
	}
	return copyFileBetween2Storages(t, srcStorage, dstStorage, srcObjPath, dstDirPath)
}

func copyFileBetween2Storages(tsk *CopyTask, srcStorage, dstStorage driver.Driver, srcFilePath, dstDirPath string) error {
	srcFile, err := op.Get(tsk.Ctx(), srcStorage, srcFilePath)
	if err != nil {
		return errors.WithMessagef(err, "failed get src [%s] file", srcFilePath)
	}
	link, _, err := op.Link(tsk.Ctx(), srcStorage, srcFilePath, model.LinkArgs{
		Header: http.Header{},
	})
	if err != nil {
		return errors.WithMessagef(err, "failed get [%s] link", srcFilePath)
	}
	fs := stream.FileStream{
		Obj: srcFile,
		Ctx: tsk.Ctx(),
	}
	// any link provided is seekable
	ss, err := stream.NewSeekableStream(fs, link)
	if err != nil {
		return errors.WithMessagef(err, "failed get [%s] stream", srcFilePath)
	}
	return op.Put(tsk.Ctx(), dstStorage, dstDirPath, ss, tsk.SetProgress, true)
}
