package common

import (
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/sign"
)

func Sign(obj model.Obj) string {
	if obj.IsDir() {
		return ""
	}
	return sign.Sign(obj.GetName())
}
