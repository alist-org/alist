package errs

import "errors"

var (
	EmptyUsername      = errors.New("username is empty")
	EmptyPassword      = errors.New("password is empty")
	WrongPassword      = errors.New("password is incorrect")
	DeleteAdminOrGuest = errors.New("cannot delete admin or guest")
)
