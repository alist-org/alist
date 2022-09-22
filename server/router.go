package server

import (
	"github.com/alist-org/alist/v3/cmd/flags"
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/message"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/alist-org/alist/v3/server/handles"
	"github.com/alist-org/alist/v3/server/middlewares"
	"github.com/alist-org/alist/v3/server/static"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Init(r *gin.Engine) {
	common.SecretKey = []byte(conf.Conf.JwtSecret)
	Cors(r)
	r.Use(middlewares.StoragesLoaded)
	WebDav(r.Group("/dav"))

	r.GET("/favicon.ico", handles.Favicon)
	r.GET("/i/:link_name", handles.Plist)
	r.GET("/d/*path", middlewares.Down, handles.Down)
	r.GET("/p/*path", middlewares.Down, handles.Proxy)

	api := r.Group("/api")
	auth := api.Group("", middlewares.Auth)

	api.POST("/auth/login", handles.Login)
	auth.GET("/me", handles.CurrentUser)
	auth.POST("/me/update", handles.UpdateCurrent)
	auth.POST("/auth/2fa/generate", handles.Generate2FA)
	auth.POST("/auth/2fa/verify", handles.Verify2FA)

	// no need auth
	public := api.Group("/public")
	public.Any("/settings", handles.PublicSettings)

	_fs(auth.Group("/fs"))
	admin(auth.Group("/admin", middlewares.AuthAdmin))
	if flags.Dev {
		dev(r.Group("/dev"))
	}
	static.Static(r)
}

func admin(g *gin.RouterGroup) {
	meta := g.Group("/meta")
	meta.GET("/list", handles.ListMetas)
	meta.GET("/get", handles.GetMeta)
	meta.POST("/create", handles.CreateMeta)
	meta.POST("/update", handles.UpdateMeta)
	meta.POST("/delete", handles.DeleteMeta)

	user := g.Group("/user")
	user.GET("/list", handles.ListUsers)
	user.GET("/get", handles.GetUser)
	user.POST("/create", handles.CreateUser)
	user.POST("/update", handles.UpdateUser)
	user.POST("/cancel_2fa", handles.Cancel2FAById)
	user.POST("/delete", handles.DeleteUser)

	storage := g.Group("/storage")
	storage.GET("/list", handles.ListStorages)
	storage.GET("/get", handles.GetStorage)
	storage.POST("/create", handles.CreateStorage)
	storage.POST("/update", handles.UpdateStorage)
	storage.POST("/delete", handles.DeleteStorage)
	storage.POST("/enable", handles.EnableStorage)
	storage.POST("/disable", handles.DisableStorage)

	driver := g.Group("/driver")
	driver.GET("/list", handles.ListDriverInfo)
	driver.GET("/names", handles.ListDriverNames)
	driver.GET("/info", handles.GetDriverInfo)

	setting := g.Group("/setting")
	setting.GET("/get", handles.GetSetting)
	setting.GET("/list", handles.ListSettings)
	setting.POST("/save", handles.SaveSettings)
	setting.POST("/delete", handles.DeleteSetting)
	setting.POST("/reset_token", handles.ResetToken)
	setting.POST("/set_aria2", handles.SetAria2)

	task := g.Group("/task")
	task.GET("/down/undone", handles.UndoneDownTask)
	task.GET("/down/done", handles.DoneDownTask)
	task.POST("/down/cancel", handles.CancelDownTask)
	task.POST("/down/delete", handles.DeleteDownTask)
	task.POST("/down/clear_done", handles.ClearDoneDownTasks)
	task.GET("/transfer/undone", handles.UndoneTransferTask)
	task.GET("/transfer/done", handles.DoneTransferTask)
	task.POST("/transfer/cancel", handles.CancelTransferTask)
	task.POST("/transfer/delete", handles.DeleteTransferTask)
	task.POST("/transfer/clear_done", handles.ClearDoneTransferTasks)
	task.GET("/upload/undone", handles.UndoneUploadTask)
	task.GET("/upload/done", handles.DoneUploadTask)
	task.POST("/upload/cancel", handles.CancelUploadTask)
	task.POST("/upload/delete", handles.DeleteUploadTask)
	task.POST("/upload/clear_done", handles.ClearDoneUploadTasks)
	task.GET("/copy/undone", handles.UndoneCopyTask)
	task.GET("/copy/done", handles.DoneCopyTask)
	task.POST("/copy/cancel", handles.CancelCopyTask)
	task.POST("/copy/delete", handles.DeleteCopyTask)
	task.POST("/copy/clear_done", handles.ClearDoneCopyTasks)

	ms := g.Group("/message")
	ms.POST("/get", message.HttpInstance.GetHandle)
	ms.POST("/send", message.HttpInstance.SendHandle)
}

func _fs(g *gin.RouterGroup) {
	g.Any("/list", handles.FsList)
	g.Any("/get", handles.FsGet)
	g.Any("/other", handles.FsOther)
	g.Any("/dirs", handles.FsDirs)
	g.POST("/mkdir", handles.FsMkdir)
	g.POST("/rename", handles.FsRename)
	g.POST("/move", handles.FsMove)
	g.POST("/copy", handles.FsCopy)
	g.POST("/remove", handles.FsRemove)
	g.PUT("/put", handles.FsStream)
	g.PUT("/form", handles.FsForm)
	g.POST("/link", middlewares.AuthAdmin, handles.Link)
	g.POST("/add_aria2", handles.AddAria2)
}

func Cors(r *gin.Engine) {
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = append(config.AllowHeaders, "Authorization", "range", "File-Path", "As-Task")
	r.Use(cors.New(config))
}
