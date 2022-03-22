package controllers

import (
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/server/common"
	"github.com/Xhofe/alist/utils"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func Down(c *gin.Context) {
	rawPath := c.Param("path")
	rawPath = utils.ParsePath(rawPath)
	log.Debugf("down: %s", rawPath)
	account, path_, driver, err := common.ParsePath(rawPath)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	if driver.Config().OnlyProxy || account.Proxy || utils.IsContain(conf.DProxyTypes, utils.Ext(rawPath)) {
		Proxy(c)
		return
	}
	link, err := driver.Link(base.Args{Path: path_, IP: c.ClientIP()}, account)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	c.Redirect(302, link.Url)
	return
}
