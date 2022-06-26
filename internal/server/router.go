package server

import (
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/server/common"
	"github.com/alist-org/alist/v3/internal/server/controllers"
	"github.com/alist-org/alist/v3/internal/server/middlewares"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Init(r *gin.Engine) {
	common.SecretKey = []byte(conf.Conf.JwtSecret)
	Cors(r)

	api := r.Group("/api", middlewares.Auth)
	api.POST("/user/login", controllers.Login)
	api.GET("/user/current", controllers.CurrentUser)
}

func Cors(r *gin.Engine) {
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = append(config.AllowHeaders, "Authorization", "range")
	r.Use(cors.New(config))
}
