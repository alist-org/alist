package errs

import (
	"errors"
	pkgerr "github.com/pkg/errors"
)

var (
	ObjectNotFound = errors.New("object not found")
	NotImplement   = errors.New("not implement")
	NotSupport     = errors.New("not support")
	RelativePath   = errors.New("access using relative path is not allowed")

	MoveBetweenTwoAccounts = errors.New("can't move files between two account, try to copy")
	UploadNotSupported     = errors.New("upload not supported")
	NotFolder              = errors.New("not a folder")
	NotFile                = errors.New("not a file")

	MetaNotFound = errors.New("meta not found")
)

func IsObjectNotFound(err error) bool {
	return pkgerr.Cause(err) == ObjectNotFound
}
