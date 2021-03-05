package v2

import (
	"github.com/Xhofe/alist/alidrive"
	"github.com/Xhofe/alist/server/controllers"
	"github.com/Xhofe/alist/server/models"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"path/filepath"
)

// get request bean
type GetReq struct {
	File     string `json:"file"`
	Password string `json:"password"`
}

// handle list request
func Get(c *gin.Context) {
	var get GetReq
	if err := c.ShouldBindJSON(&get); err != nil {
		c.JSON(200, controllers.MetaResponse(400, "Bad Request."))
		return
	}
	log.Debugf("list:%+v", get)
	dir, name := filepath.Split(get.File)
	file, err := models.GetFileByParentPathAndName(dir, name)
	if err != nil {
		c.JSON(200, controllers.MetaResponse(500, err.Error()))
		return
	}
	c.JSON(200, controllers.DataResponse(file))
}

// handle download request
func Down(c *gin.Context) {
	filePath := c.Param("file")
	log.Debugf("down:%s", filePath)
	dir, name := filepath.Split(filePath)
	fileModel, err := models.GetFileByParentPathAndName(dir, name)
	if err != nil {
		c.JSON(200, controllers.MetaResponse(500, err.Error()))
		return
	}
	file, err := alidrive.GetDownLoadUrl(fileModel.FileId)
	if err != nil {
		c.JSON(200, controllers.MetaResponse(500, err.Error()))
		return
	}
	c.Redirect(301, file.Url)
	return
}
