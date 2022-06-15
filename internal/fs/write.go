package fs

import (
	"context"

	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/operations"
	"github.com/pkg/errors"
)

func MakeDir(ctx context.Context, account driver.Driver, path string) error {
	account, actualPath, err := operations.GetAccountAndActualPath(path)
	if err != nil {
		return errors.WithMessage(err, "failed get account")
	}
	return operations.MakeDir(ctx, account, actualPath)
}

func Move(ctx context.Context, account driver.Driver, srcPath, dstPath string) error {
	srcAccount, srcActualPath, err := operations.GetAccountAndActualPath(srcPath)
	if err != nil {
		return errors.WithMessage(err, "failed get src account")
	}
	dstAccount, dstActualPath, err := operations.GetAccountAndActualPath(srcPath)
	if err != nil {
		return errors.WithMessage(err, "failed get dst account")
	}
	if srcAccount.GetAccount() != dstAccount.GetAccount() {
		return errors.WithStack(ErrMoveBetwwenTwoAccounts)
	}
	return operations.Move(ctx, account, srcActualPath, dstActualPath)
}

func Rename(ctx context.Context, account driver.Driver, srcPath, dstName string) error {
	account, srcActualPath, err := operations.GetAccountAndActualPath(srcPath)
	if err != nil {
		return errors.WithMessage(err, "failed get account")
	}
	return operations.Rename(ctx, account, srcActualPath, dstName)
}

// Copy if in an account, call move method
// if not, add copy task
func Copy(ctx context.Context, account driver.Driver, srcPath, dstPath string) error {
	srcAccount, srcActualPath, err := operations.GetAccountAndActualPath(srcPath)
	if err != nil {
		return errors.WithMessage(err, "failed get src account")
	}
	dstAccount, dstActualPath, err := operations.GetAccountAndActualPath(srcPath)
	if err != nil {
		return errors.WithMessage(err, "failed get dst account")
	}
	// copy if in an account, just call driver.Copy
	if srcAccount.GetAccount() == dstAccount.GetAccount() {
		return operations.Copy(ctx, account, srcActualPath, dstActualPath)
	}
	// not in an account
	return CopyBetween2Accounts(ctx, srcAccount, dstAccount, srcActualPath, dstActualPath)
	// srcFile, err := operations.Get(ctx, srcAccount, srcActualPath)
	// if srcFile.IsDir() {
	// 	// TODO: recursive copy
	// 	return nil
	// }
	// // TODO: add copy task, maybe like this:
	// // operations.Link(ctx,srcAccount,srcActualPath,args)
	// // get a Reader from link
	// // boxing the Reader to a driver.FileStream
	// // operations.Put(ctx,dstParentPath, stream)
	// panic("TODO")
}

func Remove(ctx context.Context, account driver.Driver, path string) error {
	account, actualPath, err := operations.GetAccountAndActualPath(path)
	if err != nil {
		return errors.WithMessage(err, "failed get account")
	}
	return operations.Remove(ctx, account, actualPath)
}

func Put(ctx context.Context, account driver.Driver, parentPath string, file model.FileStreamer) error {
	account, actualParentPath, err := operations.GetAccountAndActualPath(parentPath)
	if err != nil {
		return errors.WithMessage(err, "failed get account")
	}
	return operations.Put(ctx, account, actualParentPath, file)
}
