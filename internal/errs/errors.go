package errs

import (
	"errors"
)

var (
	NotImplement = errors.New("not implement")
	NotSupport   = errors.New("not support")
	RelativePath = errors.New("access using relative path is not allowed")

	MoveBetweenTwoStorages = errors.New("can't move files between two storages, try to copy")
	UploadNotSupported     = errors.New("upload not supported")

	MetaNotFound = errors.New("meta not found")
)
