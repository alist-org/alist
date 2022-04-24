package gowebdav

import (
	"fmt"
	"os"
)

// StatusError implements error and wraps
// an erroneous status code.
type StatusError struct {
	Status int
}

func (se StatusError) Error() string {
	return fmt.Sprintf("%d", se.Status)
}

// IsErrCode returns true if the given error
// is an os.PathError wrapping a StatusError
// with the given status code.
func IsErrCode(err error, code int) bool {
	if pe, ok := err.(*os.PathError); ok {
		se, ok := pe.Err.(StatusError)
		return ok && se.Status == code
	}
	return false
}

// IsErrNotFound is shorthand for IsErrCode
// for status 404.
func IsErrNotFound(err error) bool {
	return IsErrCode(err, 404)
}

func newPathError(op string, path string, statusCode int) error {
	return &os.PathError{
		Op:   op,
		Path: path,
		Err:  StatusError{statusCode},
	}
}

func newPathErrorErr(op string, path string, err error) error {
	return &os.PathError{
		Op:   op,
		Path: path,
		Err:  err,
	}
}
