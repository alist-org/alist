package controllers

import (
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/server/common"
	"github.com/Xhofe/alist/utils"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type LinkReq struct {
	Path string `json:"path"`
	//Password string `json:"password"`
}

// Link 返回真实的链接，且携带头,只提供给中转程序使用
func Link(c *gin.Context) {
	var req LinkReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	req.Path = utils.ParsePath(req.Path)
	rawPath := req.Path
	rawPath = utils.ParsePath(rawPath)
	log.Debugf("link: %s", rawPath)
	account, path, driver, err := common.ParsePath(rawPath)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	if driver.Config().OnlyLocal {
		common.SuccessResp(c, base.Link{
			Url: fmt.Sprintf("//%s/p%s?d=1&sign=%s", c.Request.Host, req.Path, utils.SignWithToken(utils.Base(rawPath), conf.Token)),
		})
		return
	}
	link, err := driver.Link(base.Args{Path: path, IP: c.ClientIP()}, account)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c, link)
	return
}
