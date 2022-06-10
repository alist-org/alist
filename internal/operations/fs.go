package operations

import (
	"context"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func List(ctx context.Context, path string) ([]driver.FileInfo, error) {
	account, actualPath, err := GetAccountAndActualPath(path)
	virtualFiles := GetAccountVirtualFilesByPath(path)
	if err != nil {
		if len(virtualFiles) != 0 {
			return virtualFiles, nil
		}
		return nil, errors.WithMessage(err, "failed get account")
	}
	files, err := account.List(ctx, actualPath)
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
