package model

import (
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/pkg/errors"
)

const (
	GENERAL = iota
	GUEST   // only one exists
	ADMIN
)

type User struct {
	ID       uint   `json:"id" gorm:"primaryKey"`   // unique key
	Username string `json:"username" gorm:"unique"` // username
	Password string `json:"password"`               // password
	BasePath string `json:"base_path"`              // base path
	ReadOnly bool   `json:"read_only"`              // read only
	Webdav   bool   `json:"webdav"`                 // allow webdav
	Role     int    `json:"role"`                   // user's role
}

func (u User) IsGuest() bool {
	return u.Role == GUEST
}

func (u User) IsAdmin() bool {
	return u.Role == ADMIN
}

func (u User) ValidatePassword(password string) error {
	if password == "" {
		return errors.WithStack(errs.EmptyPassword)
	}
	if u.Password != password {
		return errors.WithStack(errs.WrongPassword)
	}
	return nil
}
