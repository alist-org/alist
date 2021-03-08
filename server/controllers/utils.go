package controllers

import (
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/server/models"
	"github.com/gin-gonic/gin"
)

// handle info request
func Info(c *gin.Context) {
	c.JSON(200, DataResponse(conf.Conf.Info))
}

// rebuild tree
func RebuildTree(c *gin.Context) {
	password := c.Param("password")[1:]
	if password != conf.Conf.Server.Password {
		if password == "" {
			c.JSON(200, MetaResponse(401, "need password."))
			return
		}
		c.JSON(200, MetaResponse(401, "wrong password."))
		return
	}
	if err := models.Clear(); err != nil {
		c.JSON(200, MetaResponse(500, err.Error()))
		return
	}
	if err := models.BuildTree(); err != nil {
		c.JSON(200, MetaResponse(500, err.Error()))
		return
	}
	c.JSON(200, MetaResponse(200, "success."))
	return
}
