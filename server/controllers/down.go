package controllers

import (
	"fmt"
	"github.com/alist-org/alist/v3/internal/sign"
	stdpath "path"
	"strings"

	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/fs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/setting"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/gin-gonic/gin"
)

func Down(c *gin.Context) {
	rawPath := c.MustGet("path").(string)
	filename := stdpath.Base(rawPath)
	account, err := fs.GetAccount(rawPath)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	if shouldProxy(account, filename) {
		Proxy(c)
		return
	} else {
		link, _, err := fs.Link(c, rawPath, model.LinkArgs{
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

func Proxy(c *gin.Context) {
	rawPath := c.MustGet("path").(string)
	filename := stdpath.Base(rawPath)
	account, err := fs.GetAccount(rawPath)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	if canProxy(account, filename) {
		downProxyUrl := account.GetAccount().DownProxyUrl
		if downProxyUrl != "" {
			_, ok := c.GetQuery("d")
			if ok {
				URL := fmt.Sprintf("%s%s?sign=%s", strings.Split(downProxyUrl, "\n")[0], rawPath, sign.Sign(filename))
				c.Redirect(302, URL)
				return
			}
		}
		link, file, err := fs.Link(c, rawPath, model.LinkArgs{
			Header: c.Request.Header,
		})
		if err != nil {
			common.ErrorResp(c, err, 500)
			return
		}
		err = common.Proxy(c.Writer, c.Request, link, file)
		if err != nil {
			common.ErrorResp(c, err, 500, true)
			return
		}
	} else {
		common.ErrorStrResp(c, "proxy not allowed", 403)
		return
	}
}

// TODO need optimize
// when should be proxy?
// 1. config.MustProxy()
// 2. account.WebProxy
// 3. proxy_types
func shouldProxy(account driver.Driver, filename string) bool {
	if account.Config().MustProxy() || account.GetAccount().WebProxy {
		return true
	}
	proxyTypes := setting.GetByKey("proxy_types")
	if strings.Contains(proxyTypes, utils.Ext(filename)) {
		return true
	}
	return false
}

// TODO need optimize
// when can be proxy?
// 1. text file
// 2. config.MustProxy()
// 3. account.WebProxy
// 4. proxy_types
// solution: text_file + shouldProxy()
func canProxy(account driver.Driver, filename string) bool {
	if account.Config().MustProxy() || account.GetAccount().WebProxy {
		return true
	}
	proxyTypes := setting.GetByKey("proxy_types")
	if strings.Contains(proxyTypes, utils.Ext(filename)) {
		return true
	}
	textTypes := setting.GetByKey("text_types")
	if strings.Contains(textTypes, utils.Ext(filename)) {
		return true
	}
	return false
}
