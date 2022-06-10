package fs

import "errors"

var (
	ErrMoveBetwwenTwoAccounts = errors.New("can't move files between two account, try to copy")
)
