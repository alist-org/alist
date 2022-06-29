package controllers

import (
	"fmt"
	"github.com/alist-org/alist/v3/internal/sign"
	stdpath "path"
	"strings"
	"time"

	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/fs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type ListReq struct {
	common.PageReq
	Path     string `json:"path" form:"path"`
	Password string `json:"password" form:"password"`
}

type ObjResp struct {
	Name     string    `json:"name"`
	Size     int64     `json:"size"`
	IsDir    bool      `json:"is_dir"`
	Modified time.Time `json:"modified"`
	Sign     string    `json:"sign"`
}

type FsListResp struct {
	Content []ObjResp `json:"content"`
	Total   int64     `json:"total"`
}

func FsList(c *gin.Context) {
	var req ListReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	req.Validate()
	user := c.MustGet("user").(*model.User)
	req.Path = stdpath.Join(user.BasePath, req.Path)
	meta, err := db.GetNearestMeta(req.Path)
	if err != nil {
		if !errors.Is(errors.Cause(err), errs.MetaNotFound) {
			common.ErrorResp(c, err, 500, true)
			return
		}
	}
	c.Set("meta", meta)
	if !canAccess(user, meta, req.Path, req.Password) {
		common.ErrorStrResp(c, "password is incorrect", 401)
		return
	}
	objs, err := fs.List(c, req.Path)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	total, objs := pagination(objs, &req.PageReq)
	common.SuccessResp(c, FsListResp{
		Content: toObjResp(objs),
		Total:   int64(total),
	})
}

func canAccess(user *model.User, meta *model.Meta, path string, password string) bool {
	// if is not guest, can access
	if user.IsAdmin() || user.IgnorePassword {
		return true
	}
	// if meta is nil or password is empty, can access
	if meta == nil || meta.Password == "" {
		return true
	}
	// if meta doesn't apply to sub_folder, can access
	if !utils.PathEqual(meta.Path, path) && !meta.SubFolder {
		return true
	}
	// validate password
	return meta.Password == password
}

func pagination(objs []model.Obj, req *common.PageReq) (int, []model.Obj) {
	pageIndex, pageSize := req.PageIndex, req.PageSize
	total := len(objs)
	start := (pageIndex - 1) * pageSize
	if start > total {
		return total, []model.Obj{}
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return total, objs[start:end]
}

func toObjResp(objs []model.Obj) []ObjResp {
	var resp []ObjResp
	for _, obj := range objs {
		resp = append(resp, ObjResp{
			Name:     obj.GetName(),
			Size:     obj.GetSize(),
			IsDir:    obj.IsDir(),
			Modified: obj.ModTime(),
			Sign:     common.Sign(obj),
		})
	}
	return resp
}

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
	var rawURL string
	// obj have raw url
	if u, ok := obj.(model.URL); ok {
		rawURL = u.URL()
	} else {
		account, _ := fs.GetAccount(req.Path)
		if account.Config().MustProxy() || account.GetAccount().WebProxy {
			if account.GetAccount().DownProxyUrl != "" {
				rawURL = fmt.Sprintf("%s%s?sign=%s", strings.Split(account.GetAccount().DownProxyUrl, "\n")[0], req.Path, sign.Sign(obj.GetName()))
			} else {
				rawURL = fmt.Sprintf("%s/p%s?sign=%s", common.GetBaseUrl(c.Request), req.Path, sign.Sign(obj.GetName()))
			}
		} else {
			// if account is not proxy, use raw url by fs.Link
			link, _, err := fs.Link(c, req.Path, model.LinkArgs{})
			if err != nil {
				common.ErrorResp(c, err, 500)
				return
			}
			rawURL = link.URL
		}
	}
	common.SuccessResp(c, FsGetResp{
		ObjResp: ObjResp{
			Name:     obj.GetName(),
			Size:     obj.GetSize(),
			IsDir:    obj.IsDir(),
			Modified: obj.ModTime(),
			Sign:     common.Sign(obj),
		},
		RawURL: rawURL,
	})
}
