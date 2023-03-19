package qbittorrent

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/internal/setting"
	"github.com/alist-org/alist/v3/pkg/task"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

func AddURL(ctx context.Context, url string, dstDirPath string) error {
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
	// call qbittorrent
	id := uuid.NewString()
	tempDir := filepath.Join(conf.Conf.TempDir, "qbittorrent", id)
	err = qbclient.AddFromLink(url, tempDir, id)
	if err != nil {
		return errors.Wrapf(err, "failed to add url %s", url)
	}
	DownTaskManager.Submit(task.WithCancelCtx(&task.Task[string]{
		ID:   id,
		Name: fmt.Sprintf("download %s to [%s](%s)", url, storage.GetStorage().MountPath, dstDirActualPath),
		Func: func(tsk *task.Task[string]) error {
			m := &Monitor{
				tsk:        tsk,
				tempDir:    tempDir,
				dstDirPath: dstDirPath,
				seedtime:   setting.GetInt(conf.QbittorrentSeedtime, 0),
			}
			return m.Loop()
		},
	}))
	return nil
}
