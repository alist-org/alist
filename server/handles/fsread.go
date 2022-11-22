package handles

import (
	"fmt"
	stdpath "path"
	"strings"
	"time"

	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/fs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/sign"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type ListReq struct {
	common.PageReq
	Path     string `json:"path" form:"path"`
	Password string `json:"password" form:"password"`
	Refresh  bool   `json:"refresh"`
}

type DirReq struct {
	Path      string `json:"path" form:"path"`
	Password  string `json:"password" form:"password"`
	ForceRoot bool   `json:"force_root" form:"force_root"`
}

type ObjResp struct {
	Name     string    `json:"name"`
	Size     int64     `json:"size"`
	IsDir    bool      `json:"is_dir"`
	Modified time.Time `json:"modified"`
	Sign     string    `json:"sign"`
	Thumb    string    `json:"thumb"`
	Type     int       `json:"type"`
}

type FsListResp struct {
	Content  []ObjResp `json:"content"`
	Total    int64     `json:"total"`
	Readme   string    `json:"readme"`
	Write    bool      `json:"write"`
	Provider string    `json:"provider"`
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
	if !common.CanAccess(user, meta, req.Path, req.Password) {
		common.ErrorStrResp(c, "password is incorrect", 403)
		return
	}
	if !user.CanWrite() && !common.CanWrite(meta, req.Path) && req.Refresh {
		common.ErrorStrResp(c, "Refresh without permission", 403)
		return
	}
	objs, err := fs.List(c, req.Path, req.Refresh)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	total, objs := pagination(objs, &req.PageReq)
	provider := "unknown"
	storage, err := fs.GetStorage(req.Path)
	if err == nil {
		provider = storage.GetStorage().Driver
	}
	common.SuccessResp(c, FsListResp{
		Content:  toObjResp(objs, req.Path, isEncrypt(meta, req.Path)),
		Total:    int64(total),
		Readme:   getReadme(meta, req.Path),
		Write:    user.CanWrite() || common.CanWrite(meta, req.Path),
		Provider: provider,
	})
}

func FsDirs(c *gin.Context) {
	var req DirReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	user := c.MustGet("user").(*model.User)
	if req.ForceRoot {
		if !user.IsAdmin() {
			common.ErrorStrResp(c, "Permission denied", 403)
			return
		}
	} else {
		req.Path = stdpath.Join(user.BasePath, req.Path)
	}
	meta, err := db.GetNearestMeta(req.Path)
	if err != nil {
		if !errors.Is(errors.Cause(err), errs.MetaNotFound) {
			common.ErrorResp(c, err, 500, true)
			return
		}
	}
	c.Set("meta", meta)
	if !common.CanAccess(user, meta, req.Path, req.Password) {
		common.ErrorStrResp(c, "password is incorrect", 403)
		return
	}
	objs, err := fs.List(c, req.Path)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	dirs := filterDirs(objs)
	common.SuccessResp(c, dirs)
}

type DirResp struct {
	Name     string    `json:"name"`
	Modified time.Time `json:"modified"`
}

func filterDirs(objs []model.Obj) []DirResp {
	var dirs []DirResp
	for _, obj := range objs {
		if obj.IsDir() {
			dirs = append(dirs, DirResp{
				Name:     utils.MappingName(obj.GetName(), conf.FilenameCharMap),
				Modified: obj.ModTime(),
			})
		}
	}
	return dirs
}

func getReadme(meta *model.Meta, path string) string {
	if meta != nil && (utils.PathEqual(meta.Path, path) || meta.RSub) {
		return meta.Readme
	}
	return ""
}

func isEncrypt(meta *model.Meta, path string) bool {
	if meta == nil || meta.Password == "" {
		return false
	}
	if !utils.PathEqual(meta.Path, path) && !meta.PSub {
		return false
	}
	return true
}

func pagination(objs []model.Obj, req *common.PageReq) (int, []model.Obj) {
	pageIndex, pageSize := req.Page, req.PerPage
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

func toObjResp(objs []model.Obj, parent string, encrypt bool) []ObjResp {
	var resp []ObjResp
	for _, obj := range objs {
		thumb := ""
		if t, ok := obj.(model.Thumb); ok {
			thumb = t.Thumb()
		}
		tp := conf.FOLDER
		if !obj.IsDir() {
			tp = utils.GetFileType(obj.GetName())
		}
		resp = append(resp, ObjResp{
			Name:     utils.MappingName(obj.GetName(), conf.FilenameCharMap),
			Size:     obj.GetSize(),
			IsDir:    obj.IsDir(),
			Modified: obj.ModTime(),
			Sign:     common.Sign(obj, parent, encrypt),
			Thumb:    thumb,
			Type:     tp,
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
	RawURL   string    `json:"raw_url"`
	Readme   string    `json:"readme"`
	Provider string    `json:"provider"`
	Related  []ObjResp `json:"related"`
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
	if !common.CanAccess(user, meta, req.Path, req.Password) {
		common.ErrorStrResp(c, "password is incorrect", 403)
		return
	}
	obj, err := fs.Get(c, req.Path)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	var rawURL string

	storage, err := fs.GetStorage(req.Path)
	provider := "unknown"
	if err == nil {
		provider = storage.Config().Name
	}
	if !obj.IsDir() {
		if err != nil {
			common.ErrorResp(c, err, 500)
			return
		}
		if storage.Config().MustProxy() || storage.GetStorage().WebProxy {
			if storage.GetStorage().DownProxyUrl != "" {
				rawURL = fmt.Sprintf("%s%s?sign=%s",
					strings.Split(storage.GetStorage().DownProxyUrl, "\n")[0],
					utils.EncodePath(req.Path, true),
					sign.Sign(req.Path))
			} else {
				rawURL = fmt.Sprintf("%s/p%s?sign=%s",
					common.GetApiUrl(c.Request),
					utils.EncodePath(req.Path, true),
					sign.Sign(req.Path))
			}
		} else {
			// file have raw url
			if u, ok := obj.(model.URL); ok {
				rawURL = u.URL()
			} else {
				// if storage is not proxy, use raw url by fs.Link
				link, _, err := fs.Link(c, req.Path, model.LinkArgs{IP: c.ClientIP(), Header: c.Request.Header})
				if err != nil {
					common.ErrorResp(c, err, 500)
					return
				}
				rawURL = link.URL
			}
		}
	}
	var related []model.Obj
	parentPath := stdpath.Dir(req.Path)
	sameLevelFiles, err := fs.List(c, parentPath)
	if err == nil {
		related = filterRelated(sameLevelFiles, obj)
	}
	parentMeta, _ := db.GetNearestMeta(parentPath)
	common.SuccessResp(c, FsGetResp{
		ObjResp: ObjResp{
			Name:     utils.MappingName(obj.GetName(), conf.FilenameCharMap),
			Size:     obj.GetSize(),
			IsDir:    obj.IsDir(),
			Modified: obj.ModTime(),
			Sign:     common.Sign(obj, parentPath, isEncrypt(meta, req.Path)),
			Type:     utils.GetFileType(obj.GetName()),
		},
		RawURL:   rawURL,
		Readme:   getReadme(meta, req.Path),
		Provider: provider,
		Related:  toObjResp(related, parentPath, isEncrypt(parentMeta, parentPath)),
	})
}

func filterRelated(objs []model.Obj, obj model.Obj) []model.Obj {
	var related []model.Obj
	nameWithoutExt := strings.TrimSuffix(obj.GetName(), stdpath.Ext(obj.GetName()))
	for _, o := range objs {
		if o.GetName() == obj.GetName() {
			continue
		}
		if strings.HasPrefix(o.GetName(), nameWithoutExt) {
			related = append(related, o)
		}
	}
	return related
}

type FsOtherReq struct {
	model.FsOtherArgs
	Password string `json:"password" form:"password"`
}

func FsOther(c *gin.Context) {
	var req FsOtherReq
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
	if !common.CanAccess(user, meta, req.Path, req.Password) {
		common.ErrorStrResp(c, "password is incorrect", 403)
		return
	}
	res, err := fs.Other(c, req.FsOtherArgs)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c, res)
}
