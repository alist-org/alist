package server

import (
	"github.com/alist-org/alist/v3/server/middlewares"
	"github.com/gin-gonic/gin"
)

func dev(g *gin.RouterGroup) {
	g.GET("/path/*path", middlewares.Down, func(ctx *gin.Context) {
		rawPath := ctx.MustGet("path").(string)
		ctx.JSON(200, gin.H{
			"path": rawPath,
		})
	})
}
