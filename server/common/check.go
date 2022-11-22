package common

import (
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
)

func CanWrite(meta *model.Meta, path string) bool {
	if meta == nil || !meta.Write {
		return false
	}
	return meta.WSub || meta.Path == path
}

func CanAccess(user *model.User, meta *model.Meta, path string, password string) bool {
	// if is not guest, can access
	if user.CanAccessWithoutPassword() {
		return true
	}
	// if meta is nil or password is empty, can access
	if meta == nil || meta.Password == "" {
		return true
	}
	// if meta doesn't apply to sub_folder, can access
	if !utils.PathEqual(meta.Path, path) && !meta.PSub {
		return true
	}
	// validate password
	return meta.Password == password
}
