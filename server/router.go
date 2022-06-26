package server

import (
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/server/common"
	controllers2 "github.com/alist-org/alist/v3/server/controllers"
	"github.com/alist-org/alist/v3/server/middlewares"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Init(r *gin.Engine) {
	common.SecretKey = []byte(conf.Conf.JwtSecret)
	Cors(r)

	api := r.Group("/api", middlewares.Auth)
	api.POST("/auth/login", controllers2.Login)
	api.GET("/auth/current", controllers2.CurrentUser)

	admin := api.Group("/admin", middlewares.AuthAdmin)

	meta := admin.Group("/meta")
	meta.GET("/list", controllers2.ListMetas)
	meta.POST("/create", controllers2.CreateMeta)
	meta.POST("/update", controllers2.UpdateMeta)
	meta.POST("/delete", controllers2.DeleteMeta)
}

func Cors(r *gin.Engine) {
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = append(config.AllowHeaders, "Authorization", "range")
	r.Use(cors.New(config))
}
