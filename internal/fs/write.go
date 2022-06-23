package fs

import (
	"context"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/operations"
	"github.com/pkg/errors"
)

func makeDir(ctx context.Context, path string) error {
	account, actualPath, err := operations.GetAccountAndActualPath(path)
	if err != nil {
		return errors.WithMessage(err, "failed get account")
	}
	return operations.MakeDir(ctx, account, actualPath)
}

func move(ctx context.Context, srcPath, dstDirPath string) error {
	srcAccount, srcActualPath, err := operations.GetAccountAndActualPath(srcPath)
	if err != nil {
		return errors.WithMessage(err, "failed get src account")
	}
	dstAccount, dstDirActualPath, err := operations.GetAccountAndActualPath(dstDirPath)
	if err != nil {
		return errors.WithMessage(err, "failed get dst account")
	}
	if srcAccount.GetAccount() != dstAccount.GetAccount() {
		return errors.WithStack(errs.MoveBetweenTwoAccounts)
	}
	return operations.Move(ctx, srcAccount, srcActualPath, dstDirActualPath)
}

func rename(ctx context.Context, srcPath, dstName string) error {
	account, srcActualPath, err := operations.GetAccountAndActualPath(srcPath)
	if err != nil {
		return errors.WithMessage(err, "failed get account")
	}
	return operations.Rename(ctx, account, srcActualPath, dstName)
}

func remove(ctx context.Context, path string) error {
	account, actualPath, err := operations.GetAccountAndActualPath(path)
	if err != nil {
		return errors.WithMessage(err, "failed get account")
	}
	return operations.Remove(ctx, account, actualPath)
}
