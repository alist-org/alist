package controllers

import (
	"github.com/Xhofe/alist/alidrive"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/server/models"
	"github.com/Xhofe/alist/utils"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http/httputil"
	url2 "net/url"
	"path/filepath"
	"strings"
)

type ProxyReq struct {
	Password string `form:"pw"`
}

// Down handle download request
func Proxy(c *gin.Context) {
	if !conf.Conf.Server.Download {
		c.JSON(200, MetaResponse(403, "不允许下载或预览."))
		return
	}
	filePath := c.Param("path")[1:]
	if !utils.HasSuffixes(filePath, conf.AllowProxies) {
		c.JSON(200, MetaResponse(403, "该类型文件不允许代理."))
		return
	}
	var down ProxyReq
	if err := c.ShouldBindQuery(&down); err != nil {
		c.JSON(200, MetaResponse(400, "错误的请求."))
		return
	}
	log.Debugf("down:%s", filePath)
	dir, name := filepath.Split(filePath)
	fileModel, err := models.GetFileByDirAndName(dir, name)
	if err != nil {
		if fileModel == nil {
			c.JSON(200, MetaResponse(404, "文件不存在."))
			return
		}
		c.JSON(200, MetaResponse(500, err.Error()))
		return
	}
	if fileModel.Password != "" && down.Password != utils.Get16MD5Encode(fileModel.Password) {
		if down.Password == "" {
			c.JSON(200, MetaResponse(401, "需要密码."))
		} else {
			c.JSON(200, MetaResponse(401, "密码错误."))
		}
		return
	}
	if fileModel.Type == "folder" {
		c.JSON(200, MetaResponse(406, "无法代理目录."))
		return
	}
	drive := utils.GetDriveByName(strings.Split(filePath, "/")[0])
	if drive == nil {
		c.JSON(200, MetaResponse(500, "找不到drive."))
		return
	}
	file, err := alidrive.GetDownLoadUrl(fileModel.FileId, drive)
	if err != nil {
		c.JSON(200, MetaResponse(500, err.Error()))
		return
	}
	url, err := url2.Parse(file.Url)
	if err != nil {
		c.JSON(200, MetaResponse(500, err.Error()))
		return
	}
	target, err := url2.Parse("https://" + url.Host)
	if err != nil {
		c.JSON(200, MetaResponse(500, err.Error()))
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	c.Request.URL = url
	c.Request.Host = url.Host
	c.Request.Header.Del("Origin")
	proxy.ServeHTTP(c.Writer, c.Request)
}
