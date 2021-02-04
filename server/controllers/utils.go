package controllers

import (
	"github.com/Xhofe/alist/conf"
	"github.com/gin-gonic/gin"
)

// handle info request
func Info(c *gin.Context) {
	c.JSON(200, DataResponse(conf.Conf.Info))
}

// handle refresh_cache request
func RefreshCache(c *gin.Context) {
	password:=c.Param("password")
	if conf.Conf.Cache.Enable {
		if password == conf.Conf.Cache.RefreshPassword {
			conf.Cache.Flush()
			c.JSON(200, MetaResponse(200,"flush success."))
			return
		}
		c.JSON(200, MetaResponse(401,"wrong password."))
		return
	}
	c.JSON(200, MetaResponse(400,"disabled cache."))
	return
}