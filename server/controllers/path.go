package controllers

import (
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/server/common"
	"github.com/Xhofe/alist/utils"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"strings"
)

func Path(c *gin.Context) {
	reqV, _ := c.Get("req")
	req := reqV.(common.PathReq)
	if model.AccountsCount() > 1 && req.Path == "/" {
		files, err := model.GetAccountFiles()
		if err != nil {
			common.ErrorResp(c, err, 500)
			return
		}
		c.JSON(200, common.Resp{
			Code:    200,
			Message: "folder",
			Data:    files,
		})
		return
	}
	account, path, driver, err := common.ParsePath(req.Path)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	file, files, err := driver.Path(path, account)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	if file != nil {
		// 对于中转文件或只能中转,将链接修改为中转链接
		if driver.Config().OnlyProxy || account.Proxy {
			if account.ProxyUrl != "" {
				file.Url = fmt.Sprintf("%s%s?sign=%s", account.ProxyUrl, req.Path, utils.SignWithToken(file.Name, conf.Token))
			}else {
				file.Url = fmt.Sprintf("//%s/d%s", c.Request.Host, req.Path)
			}
		}
		c.JSON(200, common.Resp{
			Code:    200,
			Message: "file",
			Data:    []*model.File{file},
		})
	} else {
		meta, _ := model.GetMetaByPath(req.Path)
		if meta != nil && meta.Hide != "" {
			tmpFiles := make([]model.File, 0)
			hideFiles := strings.Split(meta.Hide, ",")
			for _, item := range files {
				if !utils.IsContain(hideFiles, item.Name) {
					tmpFiles = append(tmpFiles, item)
				}
			}
			files = tmpFiles
		}
		c.JSON(200, common.Resp{
			Code:    200,
			Message: "folder",
			Data:    files,
		})
	}
}

func Link(c *gin.Context) {
	reqV, _ := c.Get("req")
	req := reqV.(common.PathReq)
	rawPath := req.Path
	rawPath = utils.ParsePath(rawPath)
	log.Debugf("link: %s", rawPath)
	account, path, driver, err := common.ParsePath(rawPath)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	link, err := driver.Link(path, account)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	if driver.Config().NeedHeader {
		common.SuccessResp(c, gin.H{
			"url": link,
			"header": gin.H{
				"name":  "Authorization",
				"value": "Bearer " + account.AccessToken,
			},
		})
		return
	}
	if driver.Config().OnlyProxy {
		common.SuccessResp(c, gin.H{
			"url": fmt.Sprintf("//%s/d%s", c.Request.Host, req.Path),
		})
		return
	} else {
		common.SuccessResp(c, gin.H{
			"url": link,
		})
		return
	}
}

func Preview(c *gin.Context) {
	reqV, _ := c.Get("req")
	req := reqV.(common.PathReq)
	rawPath := req.Path
	rawPath = utils.ParsePath(rawPath)
	log.Debugf("preview: %s", rawPath)
	account, path, driver, err := common.ParsePath(rawPath)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	data, err := driver.Preview(path, account)
	if err != nil {
		common.ErrorResp(c, err, 500)
	} else {
		common.SuccessResp(c, data)
	}
}
