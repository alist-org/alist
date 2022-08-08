package server

import (
	"github.com/alist-org/alist/v3/server/common"
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
	g.GET("/hide_privacy", func(ctx *gin.Context) {
		common.ErrorStrResp(ctx, "This is ip: 1.1.1.1", 400)
	})
}
