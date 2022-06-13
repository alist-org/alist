package fs

import (
	"context"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/operations"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	stdpath "path"
	"time"
)

// List files
// TODO: hide
// TODO: sort
func List(ctx context.Context, path string) ([]driver.FileInfo, error) {
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

func Get(ctx context.Context, path string) (driver.FileInfo, error) {
	path = utils.StandardizationPath(path)
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
			return model.File{
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

func Link(ctx context.Context, path string, args driver.LinkArgs) (*driver.Link, error) {
	account, actualPath, err := operations.GetAccountAndActualPath(path)
	if err != nil {
		return nil, errors.WithMessage(err, "failed get account")
	}
	return operations.Link(ctx, account, actualPath, args)
}
