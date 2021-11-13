package server

import (
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/utils"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/url"
	"path/filepath"
)

func Down(c *gin.Context) {
	rawPath, err := url.PathUnescape(c.Param("path"))
	if err != nil {
		ErrorResp(c, err, 500)
		return
	}
	rawPath = utils.ParsePath(rawPath)
	log.Debugf("down: %s", rawPath)
	account, path, driver, err := ParsePath(rawPath)
	if err != nil {
		ErrorResp(c, err, 500)
		return
	}
	link, err := driver.Link(path, account)
	if err != nil {
		ErrorResp(c, err, 500)
		return
	}
	if account.Type == "Native" {
		c.File(link)
		return
	} else {
		c.Redirect(302, link)
		return
	}
}

func Proxy(c *gin.Context) {
	rawPath, err := url.PathUnescape(c.Param("path"))
	if err != nil {
		ErrorResp(c, err, 500)
		return
	}
	rawPath = utils.ParsePath(rawPath)
	log.Debugf("proxy: %s", rawPath)
	account, path, driver, err := ParsePath(rawPath)
	if err != nil {
		ErrorResp(c, err, 500)
		return
	}
	if !account.Proxy && utils.GetFileType(filepath.Ext(rawPath)) != conf.TEXT {
		ErrorResp(c, fmt.Errorf("[%s] not allowed proxy", account.Name), 403)
		return
	}
	link, err := driver.Link(path, account)
	if err != nil {
		ErrorResp(c, err, 500)
		return
	}
	if account.Type == "Native" {
		c.File(link)
		return
	} else {
		driver.Proxy(c)
		// TODO
	}
}
