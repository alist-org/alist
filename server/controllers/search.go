package controllers

import (
	"github.com/Xhofe/alist/server/models"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type SearchReq struct {
	Keyword string `json:"keyword" binding:"required"`
	Dir     string `json:"dir" binding:"required"`
}

func LocalSearch(c *gin.Context) {
	var req SearchReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, MetaResponse(400, "Bad Request:"+err.Error()))
		return
	}
	log.Debugf("list:%+v", req)
	files, err := models.SearchByNameInDir(req.Keyword, req.Dir)
	if err != nil {
		if files == nil {
			c.JSON(200, MetaResponse(404, "Path not found."))
			return
		}
		c.JSON(200, MetaResponse(500, err.Error()))
		return
	}
	c.JSON(200, DataResponse(files))
}

func GlobalSearch(c *gin.Context) {

}
