package driver

import "errors"

var (
	ErrorDirNotFound    = errors.New("directory not found")
	ErrorObjectNotFound = errors.New("object not found")
	ErrNotImplement     = errors.New("not implement")
	ErrNotSupport       = errors.New("not support")
	ErrRelativePath     = errors.New("access using relative path is not allowed")
)
