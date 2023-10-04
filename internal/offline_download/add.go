package offline_download

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

type AddURIArgs struct {
	URI        string
	DstDirPath string
	Tool       string
}

func AddURI(ctx context.Context, args *AddURIArgs) error {
	// get tool
	tool, err := Tools.Get(args.Tool)
	if err != nil {
		return errors.Wrapf(err, "failed get tool")
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
	gid, err := tool.AddURI(&AddUriArgs{
		Uri:     args.URI,
		UID:     uid,
		TempDir: tempDir,
		Signal:  signal,
	})
	if err != nil {
		return errors.Wrapf(err, "failed to add uri %s", args.URI)
	}
	DownTaskManager.Submit(task.WithCancelCtx(&task.Task[string]{
		ID:   gid,
		Name: fmt.Sprintf("download %s to [%s](%s)", args.URI, storage.GetStorage().MountPath, dstDirActualPath),
		Func: func(tsk *task.Task[string]) error {
			m := &Monitor{
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
