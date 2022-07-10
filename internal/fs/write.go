package fs

import (
	"context"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/operations"
	"github.com/pkg/errors"
)

func makeDir(ctx context.Context, path string) error {
	storage, actualPath, err := operations.GetStorageAndActualPath(path)
	if err != nil {
		return errors.WithMessage(err, "failed get storage")
	}
	return operations.MakeDir(ctx, storage, actualPath)
}

func move(ctx context.Context, srcPath, dstDirPath string) error {
	srcStorage, srcActualPath, err := operations.GetStorageAndActualPath(srcPath)
	if err != nil {
		return errors.WithMessage(err, "failed get src storage")
	}
	dstStorage, dstDirActualPath, err := operations.GetStorageAndActualPath(dstDirPath)
	if err != nil {
		return errors.WithMessage(err, "failed get dst storage")
	}
	if srcStorage.GetStorage() != dstStorage.GetStorage() {
		return errors.WithStack(errs.MoveBetweenTwoStorages)
	}
	return operations.Move(ctx, srcStorage, srcActualPath, dstDirActualPath)
}

func rename(ctx context.Context, srcPath, dstName string) error {
	storage, srcActualPath, err := operations.GetStorageAndActualPath(srcPath)
	if err != nil {
		return errors.WithMessage(err, "failed get storage")
	}
	return operations.Rename(ctx, storage, srcActualPath, dstName)
}

func remove(ctx context.Context, path string) error {
	storage, actualPath, err := operations.GetStorageAndActualPath(path)
	if err != nil {
		return errors.WithMessage(err, "failed get storage")
	}
	return operations.Remove(ctx, storage, actualPath)
}
