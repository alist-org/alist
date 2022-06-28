package controllers

import (
	stdpath "path"

	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/fs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
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
	meta, err := db.GetNearestMeta(req.Path)
	if err != nil {
		if !errors.Is(errors.Cause(err), errs.MetaNotFound) {
			common.ErrorResp(c, err, 500)
			return
		}
	}
	c.Set("meta", meta)
	if !canAccess(user, meta, req.Path, req.Password) {
		common.ErrorStrResp(c, "password is incorrect", 401)
		return
	}
	obj, err := fs.Get(c, req.Path)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c, FsGetResp{
		ObjResp: ObjResp{
			Name:     obj.GetName(),
			Size:     obj.GetSize(),
			IsDir:    obj.IsDir(),
			Modified: obj.ModTime(),
			Sign:     Sign(obj),
		},
		// TODO: set raw url
	})
}
