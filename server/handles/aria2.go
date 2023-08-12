package handles

import (
	"github.com/alist-org/alist/v3/internal/aria2"
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/gin-gonic/gin"
)

type SetAria2Req struct {
	Uri    string `json:"uri" form:"uri"`
	Secret string `json:"secret" form:"secret"`
}

func SetAria2(c *gin.Context) {
	var req SetAria2Req
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	items := []model.SettingItem{
		{Key: conf.Aria2Uri, Value: req.Uri, Type: conf.TypeString, Group: model.ARIA2, Flag: model.PRIVATE},
		{Key: conf.Aria2Secret, Value: req.Secret, Type: conf.TypeString, Group: model.ARIA2, Flag: model.PRIVATE},
	}
	if err := op.SaveSettingItems(items); err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	version, err := aria2.InitClient(2)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c, version)
}

type AddAria2Req struct {
	Urls []string `json:"urls"`
	Path string   `json:"path"`
}

func AddAria2(c *gin.Context) {
	user := c.MustGet("user").(*model.User)
	if !user.CanAddAria2Tasks() {
		common.ErrorStrResp(c, "permission denied", 403)
		return
	}
	if !aria2.IsAria2Ready() {
		// try to init client
		_, err := aria2.InitClient(2)
		if err != nil {
			common.ErrorResp(c, err, 500)
			return
		}
		if !aria2.IsAria2Ready() {
			common.ErrorStrResp(c, "aria2 still not ready after init", 500)
			return
		}
	}
	var req AddAria2Req
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
		err := aria2.AddURI(c, url, reqPath)
		if err != nil {
			common.ErrorResp(c, err, 500)
			return
		}
	}
	common.SuccessResp(c)
}
