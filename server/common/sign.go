package common

import (
	stdpath "path"

	"github.com/alist-org/alist/v3/internal2/conf"
	"github.com/alist-org/alist/v3/internal2/model"
	"github.com/alist-org/alist/v3/internal2/setting"
	"github.com/alist-org/alist/v3/internal2/sign"
)

func Sign(obj model.Obj, parent string, encrypt bool) string {
	if obj.IsDir() || (!encrypt && !setting.GetBool(conf.SignAll)) {
		return ""
	}
	return sign.Sign(stdpath.Join(parent, obj.GetName()))
}
