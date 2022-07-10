package fs

import (
	"context"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/operations"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/pkg/errors"
	stdpath "path"
	"time"
)

func get(ctx context.Context, path string) (model.Obj, error) {
	path = utils.StandardizePath(path)
	// maybe a virtual file
	if path != "/" {
		virtualFiles := operations.GetStorageVirtualFilesByPath(stdpath.Dir(path))
		for _, f := range virtualFiles {
			if f.GetName() == stdpath.Base(path) {
				return f, nil
			}
		}
	}
	storage, actualPath, err := operations.GetStorageAndActualPath(path)
	if err != nil {
		// if there are no storage prefix with path, maybe root folder
		if path == "/" {
			return model.Object{
				Name:     "root",
				Size:     0,
				Modified: time.Time{},
				IsFolder: true,
			}, nil
		}
		return nil, errors.WithMessage(err, "failed get storage")
	}
	return operations.Get(ctx, storage, actualPath)
}
