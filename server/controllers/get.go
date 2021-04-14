package controllers

import (
	"github.com/Xhofe/alist/alidrive"
	"github.com/Xhofe/alist/server/models"
	"github.com/Xhofe/alist/utils"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"path/filepath"
	"strings"
)

// get request bean
type GetReq struct {
	Path     string `json:"path" binding:"required"`
	Password string `json:"password"`
}

// handle get request
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
	drive := utils.GetDriveByName(strings.Split(get.Path, "/")[0])
	if drive == nil {
		c.JSON(200, MetaResponse(500, "找不到drive."))
		return
	}
	down, err := alidrive.GetDownLoadUrl(file.FileId, drive)
	if err != nil {
		c.JSON(200, MetaResponse(500, err.Error()))
		return
	}
	c.JSON(200, DataResponse(down))
}
