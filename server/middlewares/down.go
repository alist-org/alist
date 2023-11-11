package middlewares

import (
	"strings"

	"github.com/alist-org/alist/v3/internal2/conf"
	"github.com/alist-org/alist/v3/internal2/setting"

	"github.com/alist-org/alist/v3/internal2/errs"
	"github.com/alist-org/alist/v3/internal2/model"
	"github.com/alist-org/alist/v3/internal2/op"
	"github.com/alist-org/alist/v3/internal2/sign"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func Down(c *gin.Context) {
	rawPath := parsePath(c.Param("path"))
	c.Set("path", rawPath)
	meta, err := op.GetNearestMeta(rawPath)
	if err != nil {
		if !errors.Is(errors.Cause(err), errs.MetaNotFound) {
			common.ErrorResp(c, err, 500, true)
			return
		}
	}
	c.Set("meta", meta)
	// verify sign
	if needSign(meta, rawPath) {
		s := c.Query("sign")
		err = sign.Verify(rawPath, strings.TrimSuffix(s, "/"))
		if err != nil {
			common.ErrorResp(c, err, 401)
			c.Abort()
			return
		}
	}
	c.Next()
}

// TODO: implement
// path maybe contains # ? etc.
func parsePath(path string) string {
	return utils.FixAndCleanPath(path)
}

func needSign(meta *model.Meta, path string) bool {
	if setting.GetBool(conf.SignAll) {
		return true
	}
	if common.IsStorageSignEnabled(path) {
		return true
	}
	if meta == nil || meta.Password == "" {
		return false
	}
	if !meta.PSub && path != meta.Path {
		return false
	}
	return true
}
