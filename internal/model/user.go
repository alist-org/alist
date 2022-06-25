package model

import "github.com/pkg/errors"

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
	ReadOnly bool   `json:"read_only"`              // allow upload
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
		return errors.New("password is empty")
	}
	if u.Password != password {
		return errors.New("password is incorrect")
	}
	return nil
}
