package aria2

import (
	"context"
	"fmt"
	"github.com/alist-org/alist/v3/conf"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/fs"
	"github.com/alist-org/alist/v3/internal/operations"
	"github.com/alist-org/alist/v3/pkg/task"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"path/filepath"
)

func AddURI(ctx context.Context, uri string, dstDirPath string) error {
	// check account
	account, dstDirActualPath, err := operations.GetAccountAndActualPath(dstDirPath)
	if err != nil {
		return errors.WithMessage(err, "failed get account")
	}
	// check is it could upload
	if account.Config().NoUpload {
		return errors.WithStack(fs.ErrUploadNotSupported)
	}
	// check path is valid
	obj, err := operations.Get(ctx, account, dstDirActualPath)
	if err != nil {
		if !errors.Is(errors.Cause(err), driver.ErrorObjectNotFound) {
			return errors.WithMessage(err, "failed get object")
		}
	} else {
		if !obj.IsDir() {
			// can't add to a file
			return errors.WithStack(fs.ErrNotFolder)
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
	// TODO add to task manager
	TaskManager.Submit(task.WithCancelCtx(&task.Task[string, interface{}]{
		ID:   gid,
		Name: fmt.Sprintf("download %s to [%s](%s)", uri, account.GetAccount().VirtualPath, dstDirActualPath),
		Func: func(tsk *task.Task[string, interface{}]) error {
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
