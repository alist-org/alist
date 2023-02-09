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
	Url string `json:"url" form:"url"`
}

func SetQbittorrent(c *gin.Context) {
	var req SetQbittorrentReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	items := []model.SettingItem{
		{Key: conf.QbittorrentUrl, Value: req.Url, Type: conf.TypeString, Group: model.QBITTORRENT, Flag: model.PRIVATE},
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
