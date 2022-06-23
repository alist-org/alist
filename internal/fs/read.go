package fs

import (
	"context"
	stdpath "path"
	"time"

	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/operations"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// List files
// TODO: hide
// TODO: sort
func List(ctx context.Context, path string) ([]model.Obj, error) {
	account, actualPath, err := operations.GetAccountAndActualPath(path)
	virtualFiles := operations.GetAccountVirtualFilesByPath(path)
	if err != nil {
		if len(virtualFiles) != 0 {
			return virtualFiles, nil
		}
		return nil, errors.WithMessage(err, "failed get account")
	}
	files, err := operations.List(ctx, account, actualPath)
	if err != nil {
		log.Errorf("%+v", err)
		if len(virtualFiles) != 0 {
			return virtualFiles, nil
		}
		return nil, errors.WithMessage(err, "failed get files")
	}
	for _, accountFile := range virtualFiles {
		if !containsByName(files, accountFile) {
			files = append(files, accountFile)
		}
	}
	return files, nil
}

func Get(ctx context.Context, path string) (model.Obj, error) {
	path = utils.StandardizePath(path)
	// maybe a virtual file
	if path != "/" {
		virtualFiles := operations.GetAccountVirtualFilesByPath(stdpath.Dir(path))
		for _, f := range virtualFiles {
			if f.GetName() == stdpath.Base(path) {
				return f, nil
			}
		}
	}
	account, actualPath, err := operations.GetAccountAndActualPath(path)
	if err != nil {
		// if there are no account prefix with path, maybe root folder
		if path == "/" {
			return model.Object{
				Name:     "root",
				Size:     0,
				Modified: time.Time{},
				IsFolder: true,
			}, nil
		}
		return nil, errors.WithMessage(err, "failed get account")
	}
	return operations.Get(ctx, account, actualPath)
}

func Link(ctx context.Context, path string, args model.LinkArgs) (*model.Link, error) {
	account, actualPath, err := operations.GetAccountAndActualPath(path)
	if err != nil {
		return nil, errors.WithMessage(err, "failed get account")
	}
	return operations.Link(ctx, account, actualPath, args)
}
