package fs

import "errors"

var (
	ErrMoveBetweenTwoAccounts = errors.New("can't move files between two account, try to copy")
	ErrUploadNotSupported     = errors.New("upload not supported")
	ErrNotFolder              = errors.New("not a folder")
)
