package driver

import (
	"errors"

	pkgerr "github.com/pkg/errors"
)

var (
	ErrorObjectNotFound = errors.New("object not found")
	ErrNotImplement     = errors.New("not implement")
	ErrNotSupport       = errors.New("not support")
	ErrRelativePath     = errors.New("access using relative path is not allowed")
)

func IsErrObjectNotFound(err error) bool {
	return pkgerr.Cause(err) == ErrorObjectNotFound
}
