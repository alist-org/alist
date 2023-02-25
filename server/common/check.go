package common

import (
	"regexp"
	"strings"

	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
)

func CanWrite(meta *model.Meta, path string) bool {
	if meta == nil || !meta.Write {
		return false
	}
	return meta.WSub || meta.Path == path
}

func CanAccess(user *model.User, meta *model.Meta, reqPath string, password string) bool {
	// if is not guest and can access without password
	if user.CanAccessWithoutPassword() {
		return true
	}

	// check if reqPath is the same as meta.Path and user has write permission
	if meta != nil && meta.Path == reqPath && CanWrite(meta, reqPath) {
		return true
	}

	// if meta is nil or password is empty, can access
	if meta == nil || meta.Password == "" {
		return true
	}

	// if reqPath is in hide and user can't see hides, can't access
	if meta != nil && meta.Hide != "" && !user.CanSeeHides() {
		for _, hide := range strings.Split(meta.Hide, "\n") {
			re := regexp.MustCompile(hide)
			if re.MatchString(reqPath[len(meta.Path)+1:]) {
				return false
			}
		}
	}

	// if meta doesn't apply to sub_folder and reqPath is not the same as meta.Path, can access
	if !utils.PathEqual(meta.Path, reqPath) && !meta.PSub {
		return true
	}

	// validate password
	return meta.Password == password
}
