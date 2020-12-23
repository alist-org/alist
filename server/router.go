package server

import "github.com/gin-gonic/gin"

func InitRouter(engine *gin.Engine) {
	engine.Use(CrosHandler())
	InitApiRouter(engine)
}

func InitApiRouter(engine *gin.Engine) {
	v2:=engine.Group("/api/v2")
	{
		v2.POST("/get",Get)
		v2.POST("/list",List)
		v2.POST("/search",Search)
	}
}