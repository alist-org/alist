package controllers

import (
	"github.com/Xhofe/alist/alidrive"
	"github.com/Xhofe/alist/server/models"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"path/filepath"
)

// get request bean
type GetReq struct {
	Path     string `json:"path" binding:"required"`
	Password string `json:"password"`
}

// handle list request
func Get(c *gin.Context) {
	var get GetReq
	if err := c.ShouldBindJSON(&get); err != nil {
		c.JSON(200, MetaResponse(400, "Bad Request:"+err.Error()))
		return
	}
	log.Debugf("list:%+v", get)
	dir, name := filepath.Split(get.Path)
	file, err := models.GetFileByDirAndName(dir, name)
	if err != nil {
		if file == nil {
			c.JSON(200, MetaResponse(404, "Path not found."))
			return
		}
		c.JSON(200, MetaResponse(500, err.Error()))
		return
	}
	if file.Password != "" && file.Password != get.Password {
		if get.Password == "" {
			c.JSON(200, MetaResponse(401, "need password."))
		} else {
			c.JSON(200, MetaResponse(401, "wrong password."))
		}
		return
	}
	c.JSON(200, DataResponse(file))
}

type DownReq struct {
	Password string `form:"pw"`
}

// handle download request
func Down(c *gin.Context) {
	filePath := c.Param("path")[1:]
	var down DownReq
	if err := c.ShouldBindQuery(&down); err != nil {
		c.JSON(200, MetaResponse(400, "Bad Request."))
		return
	}
	log.Debugf("down:%s", filePath)
	dir, name := filepath.Split(filePath)
	fileModel, err := models.GetFileByDirAndName(dir, name)
	if err != nil {
		if fileModel == nil {
			c.JSON(200, MetaResponse(404, "Path not found."))
			return
		}
		c.JSON(200, MetaResponse(500, err.Error()))
		return
	}
	if fileModel.Password != "" && fileModel.Password != down.Password {
		if down.Password == "" {
			c.JSON(200, MetaResponse(401, "need password."))
		} else {
			c.JSON(200, MetaResponse(401, "wrong password."))
		}
		return
	}
	if fileModel.Type == "folder" {
		c.JSON(200, MetaResponse(406, "无法下载目录."))
		return
	}
	file, err := alidrive.GetDownLoadUrl(fileModel.FileId)
	if err != nil {
		c.JSON(200, MetaResponse(500, err.Error()))
		return
	}
	c.Redirect(301, file.Url)
	return
}
