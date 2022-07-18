package server

import (
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/message"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/alist-org/alist/v3/server/handles"
	"github.com/alist-org/alist/v3/server/middlewares"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Init(r *gin.Engine) {
	common.SecretKey = []byte(conf.Conf.JwtSecret)
	Cors(r)
	WebDav(r)

	r.GET("/d/*path", middlewares.Down, handles.Down)
	r.GET("/p/*path", middlewares.Down, handles.Proxy)

	r.POST("/api/auth/login", handles.Login)

	api := r.Group("/api", middlewares.Auth)
	api.GET("/auth/current", handles.CurrentUser)

	admin := api.Group("/admin", middlewares.AuthAdmin)

	meta := admin.Group("/meta")
	meta.GET("/list", handles.ListMetas)
	meta.POST("/create", handles.CreateMeta)
	meta.POST("/update", handles.UpdateMeta)
	meta.POST("/delete", handles.DeleteMeta)

	user := admin.Group("/user")
	user.GET("/list", handles.ListUsers)
	user.POST("/create", handles.CreateUser)
	user.POST("/update", handles.UpdateUser)
	user.POST("/delete", handles.DeleteUser)

	storage := admin.Group("/storage")
	storage.GET("/list", handles.ListStorages)
	storage.GET("/get", handles.GetStorage)
	storage.POST("/create", handles.CreateStorage)
	storage.POST("/update", handles.UpdateStorage)
	storage.POST("/delete", handles.DeleteStorage)

	driver := admin.Group("/driver")
	driver.GET("/list", handles.ListDriverItems)
	driver.GET("/names", handles.ListDriverNames)
	driver.GET("/items", handles.GetDriverItems)

	setting := admin.Group("/setting")
	setting.GET("/get", handles.GetSetting)
	setting.GET("/list", handles.ListSettings)
	setting.POST("/save", handles.SaveSettings)
	setting.POST("/delete", handles.DeleteSetting)
	setting.POST("/reset_token", handles.ResetToken)
	setting.POST("/set_aria2", handles.SetAria2)

	task := admin.Group("/task")
	task.GET("/down/undone", handles.UndoneDownTask)
	task.GET("/down/done", handles.DoneDownTask)
	task.POST("/down/cancel", handles.CancelDownTask)
	task.GET("/transfer/undone", handles.UndoneTransferTask)
	task.GET("/transfer/done", handles.DoneTransferTask)
	task.POST("/transfer/cancel", handles.CancelTransferTask)
	task.GET("/upload/undone", handles.UndoneUploadTask)
	task.GET("/upload/done", handles.DoneUploadTask)
	task.POST("/upload/cancel", handles.CancelUploadTask)
	task.GET("/copy/undone", handles.UndoneCopyTask)
	task.GET("/copy/done", handles.DoneCopyTask)
	task.POST("/copy/cancel", handles.CancelCopyTask)

	ms := admin.Group("/message")
	ms.GET("/get", message.PostInstance.GetHandle)
	ms.POST("/send", message.PostInstance.SendHandle)

	// guest can
	public := api.Group("/public")
	r.Any("/api/public/settings", handles.PublicSettings)
	//public.GET("/settings", controllers.PublicSettings)
	public.Any("/list", handles.FsList)
	public.Any("/get", handles.FsGet)
	public.Any("/dirs", handles.FsDirs)

	// gust can't
	fs := api.Group("/fs")
	fs.POST("/mkdir", handles.FsMkdir)
	fs.POST("/rename", handles.FsRename)
	fs.POST("/move", handles.FsMove)
	fs.POST("/copy", handles.FsCopy)
	fs.POST("/remove", handles.FsRemove)
	fs.POST("/put", handles.FsPut)
	fs.POST("/link", middlewares.AuthAdmin, handles.Link)
	fs.POST("/add_aria2", handles.AddAria2)
}

func Cors(r *gin.Engine) {
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = append(config.AllowHeaders, "Authorization", "range", "File-Path", "As-Task")
	r.Use(cors.New(config))
}
