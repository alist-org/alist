package controllers

import (
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/fs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/gin-gonic/gin"
	stdpath "path"
)

type FsGetReq struct {
	Path     string `json:"path" form:"path"`
	Password string `json:"password" form:"password"`
}

type FsGetResp struct {
	ObjResp
	RawURL string `json:"raw_url"`
}

func FsGet(c *gin.Context) {
	var req FsGetReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	user := c.MustGet("user").(*model.User)
	req.Path = stdpath.Join(user.BasePath, req.Path)
	meta, _ := db.GetNearestMeta(req.Path)
	c.Set("meta", meta)
	if !canAccess(user, meta, req.Path, req.Password) {
		common.ErrorStrResp(c, "password is incorrect", 401)
		return
	}
	data, err := fs.Get(c, req.Path)
	if err != nil {
		common.ErrorResp(c, err, 500, true)
		return
	}
	common.SuccessResp(c, FsGetResp{
		ObjResp: ObjResp{
			Name:     data.GetName(),
			Size:     data.GetSize(),
			IsDir:    data.IsDir(),
			Modified: data.ModTime(),
		},
		// TODO: set raw url
	})
}
