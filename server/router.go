package server

import (
	"github.com/Xhofe/alist/server/controllers"
	"github.com/Xhofe/alist/server/controllers/file"
	"github.com/Xhofe/alist/server/middlewares"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func InitApiRouter(r *gin.Engine) {

	// TODO from settings
	Cors(r)
	r.GET("/d/*path", middlewares.DownCheck, controllers.Down)
	r.GET("/p/*path", middlewares.DownCheck, controllers.Proxy)
	r.GET("/favicon.ico", controllers.Favicon)
	r.GET("/i/:data/ipa.plist", controllers.Plist)

	api := r.Group("/api")
	public := api.Group("/public")
	{
		path := public.Group("", middlewares.PathCheck, middlewares.CheckAccount)
		path.POST("/path", controllers.Path)
		path.POST("/preview", controllers.Preview)

		public.POST("/search", controllers.Search)

		//path.POST("/link",middlewares.Auth, controllers.Link)
		public.POST("/upload", file.UploadFiles)

		public.GET("/settings", controllers.GetSettingsPublic)
	}

	admin := api.Group("/admin")
	{
		admin.GET("/verify", controllers.Verify)
		admin.GET("/get_redirect_url", controllers.GetRedirectUrl)
		admin.POST("/oauth", controllers.OAuth)

		admin.Use(middlewares.Auth)
		admin.GET("/settings", controllers.GetSettings)
		admin.POST("/settings", controllers.SaveSettings)
		admin.DELETE("/setting", controllers.DeleteSetting)

		admin.POST("/account/create", controllers.CreateAccount)
		admin.POST("/account/save", controllers.SaveAccount)
		admin.GET("/accounts", controllers.GetAccounts)
		admin.DELETE("/account", controllers.DeleteAccount)
		admin.GET("/drivers", controllers.GetDrivers)
		admin.GET("/clear_cache", controllers.ClearCache)

		admin.GET("/metas", controllers.GetMetas)
		admin.POST("/meta/create", controllers.CreateMeta)
		admin.POST("/meta/save", controllers.SaveMeta)
		admin.DELETE("/meta", controllers.DeleteMeta)

		admin.POST("/link", controllers.Link)
		admin.DELETE("/files", file.DeleteFiles)
		admin.POST("/mkdir", file.Mkdir)
		admin.POST("/rename", file.Rename)
		admin.POST("/move", file.Move)
		admin.POST("/copy", file.Copy)
		admin.POST("/folder", file.Folder)
		admin.POST("/refresh", file.RefreshFolder)
	}
	WebDav(r)
	Static(r)
}

func Cors(r *gin.Engine) {
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = append(config.AllowHeaders, "Authorization", "range")
	r.Use(cors.New(config))
}
