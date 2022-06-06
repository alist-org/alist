package driver

import "errors"

var (
	ErrorDirNotFound    = errors.New("directory not found")
	ErrorObjectNotFound = errors.New("object not found")
)
