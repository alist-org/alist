package server

import (
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

func CrosHandler() gin.HandlerFunc {
	return func(context *gin.Context) {
		method := context.Request.Method
		// 设置跨域
		if conf.Conf.Info.SiteUrl=="*"||utils.ContainsString(conf.Origins,context.GetHeader("Origin"))!=-1 {
			context.Header("Access-Control-Allow-Origin",context.GetHeader("Origin"))
		}else {
			context.Header("Access-Control-Allow-Origin", conf.Conf.Info.SiteUrl)//跨域访问
		}
		context.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE,UPDATE")
		context.Header("Access-Control-Allow-Headers", "Content-Length,session,Accept, Origin, Host, Connection, Accept-Encoding, Accept-Language, Keep-Alive, User-Agent, Cache-Control, Content-Type, Pragma")
		context.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers,Cache-Control,Content-Language,Content-Type,Expires,Last-Modified,Pragma,FooBar")
		context.Header("Access-Control-Max-Age", "172800")
		context.Header("Access-Control-Allow-Credentials", "true")
		//context.Set("content-type", "application/json")  //设置返回格式是json

		if method == "OPTIONS" {
			context.JSON(http.StatusOK, gin.H{})
		}

		//处理请求
		context.Next()
	}
}