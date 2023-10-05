package tool

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/pkg/task"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type AddURLArgs struct {
	URL        string
	DstDirPath string
	Tool       string
}

func AddURL(ctx context.Context, args *AddURLArgs) error {
	// get tool
	tool, err := Tools.Get(args.Tool)
	if err != nil {
		return errors.Wrapf(err, "failed get tool")
	}
	// check tool is ready
	if !tool.IsReady() {
		// try to init tool
		if _, err := tool.Init(); err != nil {
			return errors.Wrapf(err, "failed init tool %s", args.Tool)
		}
	}
	// check storage
	storage, dstDirActualPath, err := op.GetStorageAndActualPath(args.DstDirPath)
	if err != nil {
		return errors.WithMessage(err, "failed get storage")
	}
	// check is it could upload
	if storage.Config().NoUpload {
		return errors.WithStack(errs.UploadNotSupported)
	}
	// check path is valid
	obj, err := op.Get(ctx, storage, dstDirActualPath)
	if err != nil {
		if !errs.IsObjectNotFound(err) {
			return errors.WithMessage(err, "failed get object")
		}
	} else {
		if !obj.IsDir() {
			// can't add to a file
			return errors.WithStack(errs.NotFolder)
		}
	}

	uid := uuid.NewString()
	tempDir := filepath.Join(conf.Conf.TempDir, args.Tool, uid)
	signal := make(chan int)
	gid, err := tool.AddURL(&AddUrlArgs{
		Url:     args.URL,
		UID:     uid,
		TempDir: tempDir,
		Signal:  signal,
	})
	if err != nil {
		return errors.Wrapf(err, "[%s] failed to add uri %s", args.Tool, args.URL)
	}
	DownTaskManager.Submit(task.WithCancelCtx(&task.Task[string]{
		ID:   gid,
		Name: fmt.Sprintf("download %s to [%s](%s)", args.URL, storage.GetStorage().MountPath, dstDirActualPath),
		Func: func(tsk *task.Task[string]) error {
			m := &Monitor{
				tool:       tool,
				tsk:        tsk,
				tempDir:    tempDir,
				dstDirPath: args.DstDirPath,
				signal:     signal,
			}
			return m.Loop()
		},
	}))
	return nil
}
