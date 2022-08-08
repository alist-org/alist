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
	ID       uint   `json:"id" gorm:"primaryKey"`                      // unique key
	Username string `json:"username" gorm:"unique" binding:"required"` // username
	Password string `json:"password"`                                  // password
	BasePath string `json:"base_path"`                                 // base path
	Role     int    `json:"role"`                                      // user's role
	// Determine permissions by bit
	//  0: can see hidden files
	//  1: can access without password
	//  2: can add aria2 tasks
	//  3: can mkdir and upload
	//  4: can rename
	//  5: can move
	//  6: can copy
	//  7: can remove
	//  8: webdav read
	//  9: webdav write
	Permission int32  `json:"permission"`
	OtpSecret  string `json:"-"`
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

func (u User) CanSeeHides() bool {
	return u.IsAdmin() || u.Permission&1 == 1
}

func (u User) CanAccessWithoutPassword() bool {
	return u.IsAdmin() || (u.Permission>>1)&1 == 1
}

func (u User) CanAddAria2Tasks() bool {
	return u.IsAdmin() || (u.Permission>>2)&1 == 1
}

func (u User) CanWrite() bool {
	return u.IsAdmin() || (u.Permission>>3)&1 == 1
}

func (u User) CanRename() bool {
	return u.IsAdmin() || (u.Permission>>4)&1 == 1
}

func (u User) CanMove() bool {
	return u.IsAdmin() || (u.Permission>>5)&1 == 1
}

func (u User) CanCopy() bool {
	return u.IsAdmin() || (u.Permission>>6)&1 == 1
}

func (u User) CanRemove() bool {
	return u.IsAdmin() || (u.Permission>>7)&1 == 1
}

func (u User) CanWebdavRead() bool {
	return u.IsAdmin() || (u.Permission>>8)&1 == 1
}

func (u User) CanWebdavManage() bool {
	return u.IsAdmin() || (u.Permission>>9)&1 == 1
}
