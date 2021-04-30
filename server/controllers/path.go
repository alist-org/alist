package controllers

import (
	"github.com/Xhofe/alist/server/models"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"path/filepath"
)

// path request bean
type PathReq struct {
	Path     string `json:"path" binding:"required"`
	Password string `json:"password"`
}

// handle path request
func Path(c *gin.Context) {
	var req PathReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, MetaResponse(400, "Bad Request:"+err.Error()))
		return
	}
	log.Debugf("path:%+v", req)
	// find model
	dir, name := filepath.Split(req.Path)
	file, err := models.GetFileByDirAndName(dir, name)
	if err != nil {
		// folder model not exist
		if file == nil {
			c.JSON(200, MetaResponse(404, "path not found.(第一次请先点击网页底部rebuild)"))
			return
		}
		c.JSON(200, MetaResponse(500, err.Error()))
		return
	}
	// check password
	if file.Password != "" && file.Password != req.Password {
		if req.Password == "" {
			c.JSON(200, MetaResponse(401, "need password."))
		} else {
			c.JSON(200, MetaResponse(401, "wrong password."))
		}
		return
	}
	// file
	if file.Type == "file" {
		if file.Password == "" {
			file.Password = "n"
		} else {
			file.Password = "y"
		}
		c.JSON(200, DataResponse(file))
		return
	}
	// folder
	files, err := models.GetFilesByDir(req.Path + "/")
	if err != nil {
		c.JSON(200, MetaResponse(500, err.Error()))
		return
	}
	// delete password
	for i, _ := range *files {
		if (*files)[i].Password == "" {
			(*files)[i].Password = "n"
		} else {
			(*files)[i].Password = "y"
		}
	}
	c.JSON(200, DataResponse(files))
}
