package errs

import (
	"errors"

	pkgerr "github.com/pkg/errors"
)

var (
	ObjectNotFound = errors.New("object not found")
	NotFolder      = errors.New("not a folder")
	NotFile        = errors.New("not a file")
)

func IsObjectNotFound(err error) bool {
	return errors.Is(pkgerr.Cause(err), ObjectNotFound)
}
