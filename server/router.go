package server

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func InitApiRouter(r *gin.Engine) {

	// TODO from settings
	Cors(r)
	r.GET("/d/*path", Down)
	r.GET("/p/*path", Proxy)

	api := r.Group("/api")
	public := api.Group("/public")
	{
		public.POST("/path", CheckAccount, Path)
		public.POST("/preview", CheckAccount, Preview)
		public.GET("/settings", GetSettingsPublic)
		public.POST("/link", CheckAccount, Link)
	}

	admin := api.Group("/admin")
	{
		admin.Use(Auth)
		admin.GET("/login", Login)
		admin.GET("/settings", GetSettings)
		admin.POST("/settings", SaveSettings)
		admin.POST("/account/create", CreateAccount)
		admin.POST("/account/save", SaveAccount)
		admin.GET("/accounts", GetAccounts)
		admin.DELETE("/account", DeleteAccount)
		admin.GET("/drivers", GetDrivers)
		admin.GET("/clear_cache", ClearCache)

		admin.GET("/metas", GetMetas)
		admin.POST("/meta/create", CreateMeta)
		admin.POST("/meta/save", SaveMeta)
		admin.DELETE("/meta", DeleteMeta)
	}
	Static(r)
}

func Cors(r *gin.Engine) {
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = append(config.AllowHeaders, "Authorization")
	r.Use(cors.New(config))
}
