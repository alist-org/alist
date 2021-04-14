package controllers

import (
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/server/models"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// handle info request
func Info(c *gin.Context) {
	c.JSON(200, DataResponse(conf.Conf.Info))
}

type RebuildReq struct {
	Path     string `json:"path" binding:"required"`
	Password string `json:"password"`
	Depth    int    `json:"depth"`
}

// rebuild tree
func RebuildTree(c *gin.Context) {
	var req RebuildReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, MetaResponse(400, "Bad Request:"+err.Error()))
		return
	}
	log.Debugf("rebuild:%+v", req)
	password := req.Password
	if password != conf.Conf.Server.Password {
		if password == "" {
			c.JSON(200, MetaResponse(401, "need password."))
			return
		}
		c.JSON(200, MetaResponse(401, "wrong password."))
		return
	}
	if err := models.BuildTreeWithPath(req.Path, req.Depth); err != nil {
		c.JSON(200, MetaResponse(500, err.Error()))
		return
	}
	c.JSON(200, MetaResponse(200, "success."))
	return
}
