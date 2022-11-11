package common

import (
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/sign"
	stdpath "path"
)

func Sign(obj model.Obj, parent string, encrypt bool) string {
	if obj.IsDir() || !encrypt {
		return ""
	}
	return sign.Sign(stdpath.Join(parent, obj.GetName()))
}
