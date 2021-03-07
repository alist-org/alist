package controllers

import (
	"github.com/Xhofe/alist/server/models"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"path/filepath"
)

// list request bean
type ListReq struct {
	Path     string `json:"path" binding:"required"`
	Password string `json:"password"`
}

// handle list request
func List(c *gin.Context) {
	var list ListReq
	if err := c.ShouldBindJSON(&list); err != nil {
		c.JSON(200, MetaResponse(400, "Bad Request:"+err.Error()))
		return
	}
	log.Debugf("list:%+v", list)
	// find folder model
	dir, file := filepath.Split(list.Path)
	fileModel, err := models.GetFileByParentPathAndName(dir, file)
	if err != nil {
		// folder model not exist
		if fileModel == nil {
			c.JSON(200, MetaResponse(404, "folder not found."))
			return
		}
		c.JSON(200, MetaResponse(500, err.Error()))
		return
	}
	// check password
	if fileModel.Password != "" && fileModel.Password != list.Password {
		if list.Password == "" {
			c.JSON(200, MetaResponse(401, "need password."))
		} else {
			c.JSON(200, MetaResponse(401, "wrong password."))
		}
		return
	}
	files, err := models.GetFilesByParentPath(list.Path + "/")
	// delete password
	for i, _ := range *files {
		(*files)[i].Password = ""
	}
	if err != nil {
		c.JSON(200, MetaResponse(500, err.Error()))
		return
	}
	c.JSON(200, DataResponse(files))
}
