package common

import (
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/sign"
)

func Sign(obj model.Obj, encrypt bool) string {
	if obj.IsDir() || !encrypt {
		return ""
	}
	return sign.Sign(obj.GetName())
}
