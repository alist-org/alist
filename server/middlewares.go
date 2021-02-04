package server

import (
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/server/controllers"
	"github.com/Xhofe/alist/utils"
	"github.com/gin-gonic/gin"
)

// handle cors request
func CorsHandler() gin.HandlerFunc {
	return func(context *gin.Context) {
		origin:=context.GetHeader("Origin")
		// 同源
		if origin == "" {
			context.Next()
			return
		}
		method := context.Request.Method
		// 设置跨域
		context.Header("Access-Control-Allow-Origin",origin)
		context.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE,UPDATE")
		context.Header("Access-Control-Allow-Headers", "Content-Length,session,Accept, Origin, Host, Connection, Accept-Encoding, Accept-Language, Keep-Alive, User-Agent, Cache-Control, Content-Type")
		context.Header("Access-Control-Expose-Headers", "Content-Length,Cache-Control,Content-Language,Content-Type,Expires,Last-Modified")
		context.Header("Access-Control-Max-Age", "172800")
		// 信任域名
		if conf.Conf.Server.SiteUrl!="*"&&utils.ContainsString(conf.Origins,context.GetHeader("Origin"))==-1 {
			context.JSON(200,controllers.MetaResponse(413,"The origin is not in the site_url list, please configure it correctly."))
			context.Abort()
		}
		if method == "OPTIONS" {
			context.AbortWithStatus(204)
		}
		//处理请求
		context.Next()
	}
}