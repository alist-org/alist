package errs

import (
	"errors"
	pkgerr "github.com/pkg/errors"
)

var (
	ErrorObjectNotFound = errors.New("object not found")
	ErrNotImplement     = errors.New("not implement")
	ErrNotSupport       = errors.New("not support")
	ErrRelativePath     = errors.New("access using relative path is not allowed")

	ErrMoveBetweenTwoAccounts = errors.New("can't move files between two account, try to copy")
	ErrUploadNotSupported     = errors.New("upload not supported")
	ErrNotFolder              = errors.New("not a folder")
)

func IsErrObjectNotFound(err error) bool {
	return pkgerr.Cause(err) == ErrorObjectNotFound
}
