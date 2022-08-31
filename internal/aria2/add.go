package aria2

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

func AddURI(ctx context.Context, uri string, dstDirPath string) error {
	// check storage
	storage, dstDirActualPath, err := op.GetStorageAndActualPath(dstDirPath)
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
	// call aria2 rpc
	tempDir := filepath.Join(conf.Conf.TempDir, "aria2", uuid.NewString())
	options := map[string]interface{}{
		"dir": tempDir,
	}
	gid, err := client.AddURI([]string{uri}, options)
	if err != nil {
		return errors.Wrapf(err, "failed to add uri %s", uri)
	}
	DownTaskManager.Submit(task.WithCancelCtx(&task.Task[string]{
		ID:   gid,
		Name: fmt.Sprintf("download %s to [%s](%s)", uri, storage.GetStorage().MountPath, dstDirActualPath),
		Func: func(tsk *task.Task[string]) error {
			m := &Monitor{
				tsk:        tsk,
				tempDir:    tempDir,
				retried:    0,
				dstDirPath: dstDirPath,
			}
			return m.Loop()
		},
	}))
	return nil
}
