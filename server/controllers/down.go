package controllers

import (
	stdpath "path"
	"strings"

	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/fs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/setting"
	"github.com/alist-org/alist/v3/internal/sign"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func Down(c *gin.Context) {
	rawPath := parsePath(c.Param("path"))
	filename := stdpath.Base(rawPath)
	meta, err := db.GetNearestMeta(rawPath)
	if err != nil {
		if !errors.Is(errors.Cause(err), errs.MetaNotFound) {
			common.ErrorResp(c, err, 500, true)
			return
		}
	}
	// verify sign
	if needSign(meta, rawPath) {
		s := c.Param("sign")
		err = sign.Verify(filename, s)
		if err != nil {
			common.ErrorResp(c, err, 401)
			return
		}
	}
	account, err := fs.GetAccount(rawPath)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	if needProxy(account, filename) {
		link, err := fs.Link(c, rawPath, model.LinkArgs{
			Header: c.Request.Header,
		})
		if err != nil {
			common.ErrorResp(c, err, 500)
			return
		}
		obj, err := fs.Get(c, rawPath)
		if err != nil {
			common.ErrorResp(c, err, 500)
			return
		}
		err = common.Proxy(c.Writer, c.Request, link, obj)
		if err != nil {
			common.ErrorResp(c, err, 500, true)
			return
		}
	} else {
		link, err := fs.Link(c, rawPath, model.LinkArgs{
			IP:     c.ClientIP(),
			Header: c.Request.Header,
		})
		if err != nil {
			common.ErrorResp(c, err, 500)
			return
		}
		c.Redirect(302, link.URL)
	}
}

// TODO: implement
// path maybe contains # ? etc.
func parsePath(path string) string {
	return utils.StandardizePath(path)
}

func needSign(meta *model.Meta, path string) bool {
	if meta == nil || meta.Password == "" {
		return false
	}
	if !meta.SubFolder && path != meta.Path {
		return false
	}
	return true
}

func needProxy(account driver.Driver, filename string) bool {
	config := account.Config()
	if config.MustProxy() {
		return true
	}
	proxyTypes := setting.GetByKey("proxy_types")
	if strings.Contains(proxyTypes, utils.Ext(filename)) {
		return true
	}
	return false
}
