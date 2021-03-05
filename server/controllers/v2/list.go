package v2

import (
	"github.com/Xhofe/alist/server/controllers"
	"github.com/Xhofe/alist/server/models"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// list request bean
type ListReq struct {
	Path     string `json:"path"`
	Password string `json:"password"`
}

// handle list request
func List(c *gin.Context) {
	var list ListReq
	if err := c.ShouldBindJSON(&list); err != nil {
		c.JSON(200, controllers.MetaResponse(400, "Bad Request."))
		return
	}
	log.Debugf("list:%+v", list)
	files, err := models.GetFilesByParentPath(list.Path)
	if err != nil {
		c.JSON(200, controllers.MetaResponse(500, err.Error()))
		return
	}
	c.JSON(200, controllers.DataResponse(files))
}
