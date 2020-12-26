package server

import (
	"github.com/Xhofe/alist/conf"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func InitRouter(engine *gin.Engine) {
	log.Infof("初始化路由...")
	engine.Use(CrosHandler())
	InitApiRouter(engine)
}

func InitApiRouter(engine *gin.Engine) {
	engine.Use(static.Serve("/",static.LocalFile(conf.Conf.Server.Static,false)))
	engine.NoRoute(func(c *gin.Context) {
		c.File(conf.Conf.Server.Static+"/index.html")
	})
	v2:=engine.Group("/api")
	{
		v2.GET("/info",Info)
		v2.POST("/get",Get)
		v2.POST("/list",List)
		v2.POST("/search",Search)
	}
}