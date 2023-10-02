package handles

import (
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/internal/qbittorrent"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/gin-gonic/gin"
)

type SetQbittorrentReq struct {
	Url      string `json:"url" form:"url"`
	Seedtime string `json:"seedtime" form:"seedtime"`
}

func SetQbittorrent(c *gin.Context) {
	var req SetQbittorrentReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	items := []model.SettingItem{
		{Key: conf.QbittorrentUrl, Value: req.Url, Type: conf.TypeString, Group: model.SINGLE, Flag: model.PRIVATE},
		{Key: conf.QbittorrentSeedtime, Value: req.Seedtime, Type: conf.TypeNumber, Group: model.SINGLE, Flag: model.PRIVATE},
	}
	if err := op.SaveSettingItems(items); err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	if err := qbittorrent.InitClient(); err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c, "ok")
}

type AddQbittorrentReq struct {
	Urls []string `json:"urls"`
	Path string   `json:"path"`
}

func AddQbittorrent(c *gin.Context) {
	user := c.MustGet("user").(*model.User)
	if !user.CanAddQbittorrentTasks() {
		common.ErrorStrResp(c, "permission denied", 403)
		return
	}
	if !qbittorrent.IsQbittorrentReady() {
		// try to init client
		err := qbittorrent.InitClient()
		if err != nil {
			common.ErrorResp(c, err, 500)
			return
		}
		if !qbittorrent.IsQbittorrentReady() {
			common.ErrorStrResp(c, "qbittorrent still not ready after init", 500)
			return
		}
	}
	var req AddQbittorrentReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	reqPath, err := user.JoinPath(req.Path)
	if err != nil {
		common.ErrorResp(c, err, 403)
		return
	}
	for _, url := range req.Urls {
		err := qbittorrent.AddURL(c, url, reqPath)
		if err != nil {
			common.ErrorResp(c, err, 500)
			return
		}
	}
	common.SuccessResp(c)
}
