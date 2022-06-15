package fs

import (
	"context"
	stdpath "path"

	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/operations"
	"github.com/pkg/errors"
)

func CopyBetween2Accounts(ctx context.Context, srcAccount, dstAccount driver.Driver, srcPath, dstPath string) error {
	srcFile, err := operations.Get(ctx, srcAccount, srcPath)
	if err != nil {
		return errors.WithMessagef(err, "failed get src [%s] file", srcPath)
	}
	if srcFile.IsDir() {
		files, err := operations.List(ctx, srcAccount, srcPath)
		if err != nil {
			return errors.WithMessagef(err, "failed list src [%s] files", srcPath)
		}
		for _, file := range files {
			srcFilePath := stdpath.Join(srcPath, file.GetName())
			dstFilePath := stdpath.Join(dstPath, file.GetName())
			if err := CopyBetween2Accounts(ctx, srcAccount, dstAccount, srcFilePath, dstFilePath); err != nil {
				return errors.WithMessagef(err, "failed copy file [%s] to [%s]", srcFilePath, dstFilePath)
			}
		}
	}
	link, err := operations.Link(ctx, srcAccount, srcPath, model.LinkArgs{})
	if err != nil {
		return errors.WithMessagef(err, "failed get [%s] link", srcPath)
	}
	stream, err := getFileStreamFromLink(srcFile, link)
	if err != nil {
		return errors.WithMessagef(err, "failed get [%s] stream", srcPath)
	}
	// TODO add as task
	return operations.Put(ctx, dstAccount, dstPath, stream)
}
