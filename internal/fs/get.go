package fs

import (
	"context"
	stdpath "path"
	"time"

	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/pkg/errors"
)

func get(ctx context.Context, path string) (model.Obj, error) {
	path = utils.FixAndCleanPath(path)
	// maybe a virtual file
	if path != "/" {
		virtualFiles := op.GetStorageVirtualFilesByPath(stdpath.Dir(path))
		for _, f := range virtualFiles {
			if f.GetName() == stdpath.Base(path) {
				return f, nil
			}
		}
	}
	storage, actualPath, err := op.GetStorageAndActualPath(path)
	if err != nil {
		// if there are no storage prefix with path, maybe root folder
		if path == "/" {
			return &model.Object{
				Name:     "root",
				Size:     0,
				Modified: time.Time{},
				IsFolder: true,
			}, nil
		}
		return nil, errors.WithMessage(err, "failed get storage")
	}
	return op.Get(ctx, storage, actualPath)
}
