package common

import (
	"path"
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

func IsApply(metaPath, reqPath string, applySub bool) bool {
	if utils.PathEqual(metaPath, reqPath) {
		return true
	}
	return utils.IsSubPath(metaPath, reqPath) && applySub
}

func CanAccess(user *model.User, meta *model.Meta, reqPath string, password string) bool {
	// if the reqPath is in hide (only can check the nearest meta) and user can't see hides, can't access
	if meta != nil && !user.CanSeeHides() && meta.Hide != "" &&
		IsApply(meta.Path, path.Dir(reqPath), meta.HSub) { // the meta should apply to the parent of current path
		for _, hide := range strings.Split(meta.Hide, "\n") {
			re := regexp.MustCompile(hide)
			if re.MatchString(path.Base(reqPath)) {
				return false
			}
		}
	}
	// if is not guest and can access without password
	if user.CanAccessWithoutPassword() {
		return true
	}
	// if meta is nil or password is empty, can access
	if meta == nil || meta.Password == "" {
		return true
	}
	// if meta doesn't apply to sub_folder, can access
	if !utils.PathEqual(meta.Path, reqPath) && !meta.PSub {
		return true
	}
	// validate password
	return meta.Password == password
}
