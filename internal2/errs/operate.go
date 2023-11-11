package errs

import "errors"

var (
	PermissionDenied = errors.New("permission denied")
)
