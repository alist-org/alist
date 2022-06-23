package fs

import (
	"context"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/operations"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// List files
// TODO: hide
// TODO: sort
func list(ctx context.Context, path string) ([]model.Obj, error) {
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
