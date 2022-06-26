package server

import (
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/alist-org/alist/v3/server/controllers"
	"github.com/alist-org/alist/v3/server/middlewares"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Init(r *gin.Engine) {
	common.SecretKey = []byte(conf.Conf.JwtSecret)
	Cors(r)

	api := r.Group("/api", middlewares.Auth)
	api.POST("/auth/login", controllers.Login)
	api.GET("/auth/current", controllers.CurrentUser)

	admin := api.Group("/admin", middlewares.AuthAdmin)

	meta := admin.Group("/meta")
	meta.GET("/list", controllers.ListMetas)
	meta.POST("/create", controllers.CreateMeta)
	meta.POST("/update", controllers.UpdateMeta)
	meta.POST("/delete", controllers.DeleteMeta)
}

func Cors(r *gin.Engine) {
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = append(config.AllowHeaders, "Authorization", "range")
	r.Use(cors.New(config))
}
