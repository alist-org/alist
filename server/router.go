package server

import (
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/server/controllers"
	"github.com/Xhofe/alist/server/controllers/v1"
	v2 "github.com/Xhofe/alist/server/controllers/v2"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// init router
func InitRouter(engine *gin.Engine) {
	log.Infof("初始化路由...")
	engine.Use(CorsHandler())
	engine.Use(static.Serve("/",static.LocalFile(conf.Conf.Server.Static,false)))
	engine.NoRoute(func(c *gin.Context) {
		c.File(conf.Conf.Server.Static+"/index.html")
	})
	InitApiRouter(engine)
}

// init api router
func InitApiRouter(engine *gin.Engine) {
	apiV1 :=engine.Group("/api/v1")
	{
		apiV1.GET("/info",controllers.Info)
		apiV1.POST("/get", v1.Get)
		apiV1.POST("/list", v1.List)
		apiV1.POST("/search", v1.Search)
		apiV1.POST("/office_preview", v1.OfficePreview)
		apiV1.GET("/d/*file_id", v1.Down)
	}
	apiV2:=engine.Group("/api")
	{
		apiV2.POST("/list",v2.List)
		apiV2.POST("/get",v2.Get)
	}
	engine.GET("/d/*file",v2.Down)
	engine.GET("/cache/:password",controllers.RefreshCache)
	engine.GET("/rebuild",controllers.RebuildTree)
}