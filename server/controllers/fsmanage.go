package controllers

import (
	"fmt"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/fs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/sign"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/gin-gonic/gin"
	stdpath "path"
	"strconv"
	"time"
)

type MkdirOrLinkReq struct {
	Path string `json:"path" form:"path"`
}

func FsMkdir(c *gin.Context) {
	var req MkdirOrLinkReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	user := c.MustGet("user").(*model.User)
	req.Path = stdpath.Join(user.BasePath, req.Path)
	if !user.CanWrite() {
		meta, err := db.GetNearestMeta(req.Path)
		if err != nil {
			common.ErrorResp(c, err, 500)
			return
		}
		if !canWrite(meta, req.Path) {
			common.ErrorResp(c, errs.PermissionDenied, 403)
			return
		}
	}
	if err := fs.MakeDir(c, req.Path); err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	fs.ClearCache(stdpath.Dir(req.Path))
	common.SuccessResp(c)
}

func canWrite(meta *model.Meta, path string) bool {
	if meta == nil || !meta.Write {
		return false
	}
	return meta.WSub || meta.Path == path
}

type MoveCopyReq struct {
	SrcDir string   `json:"src_dir"`
	DstDir string   `json:"dst_dir"`
	Names  []string `json:"names"`
}

func FsMove(c *gin.Context) {
	var req MoveCopyReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	if len(req.Names) == 0 {
		common.ErrorStrResp(c, "Empty file names", 400)
		return
	}
	user := c.MustGet("user").(*model.User)
	if !user.CanMove() {
		common.ErrorResp(c, errs.PermissionDenied, 403)
		return
	}
	req.SrcDir = stdpath.Join(user.BasePath, req.SrcDir)
	req.DstDir = stdpath.Join(user.BasePath, req.DstDir)
	for _, name := range req.Names {
		err := fs.Move(c, stdpath.Join(req.SrcDir, name), req.DstDir)
		if err != nil {
			common.ErrorResp(c, err, 500)
			return
		}
	}
	fs.ClearCache(req.SrcDir)
	fs.ClearCache(req.DstDir)
	common.SuccessResp(c)
}

func FsCopy(c *gin.Context) {
	var req MoveCopyReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	if len(req.Names) == 0 {
		common.ErrorStrResp(c, "Empty file names", 400)
		return
	}
	user := c.MustGet("user").(*model.User)
	if !user.CanCopy() {
		common.ErrorResp(c, errs.PermissionDenied, 403)
		return
	}
	req.SrcDir = stdpath.Join(user.BasePath, req.SrcDir)
	req.DstDir = stdpath.Join(user.BasePath, req.DstDir)
	var addedTask []string
	for _, name := range req.Names {
		ok, err := fs.Copy(c, stdpath.Join(req.SrcDir, name), req.DstDir)
		if ok {
			addedTask = append(addedTask, name)
		}
		if err != nil {
			common.ErrorResp(c, err, 500)
			return
		}
	}
	if len(req.Names) != len(addedTask) {
		fs.ClearCache(req.DstDir)
	}
	if len(addedTask) > 0 {
		common.SuccessResp(c, fmt.Sprintf("Added %d tasks", len(addedTask)))
	} else {
		common.SuccessResp(c)
	}
}

type RenameReq struct {
	Path string `json:"path"`
	Name string `json:"name"`
}

func FsRename(c *gin.Context) {
	var req RenameReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	user := c.MustGet("user").(*model.User)
	if !user.CanRename() {
		common.ErrorResp(c, errs.PermissionDenied, 403)
		return
	}
	req.Path = stdpath.Join(user.BasePath, req.Path)
	if err := fs.Rename(c, req.Path, req.Name); err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	fs.ClearCache(stdpath.Dir(req.Path))
	common.SuccessResp(c)
}

type RemoveReq struct {
	Dir   string   `json:"dir"`
	Names []string `json:"names"`
}

func FsRemove(c *gin.Context) {
	var req RemoveReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	if len(req.Names) == 0 {
		common.ErrorStrResp(c, "Empty file names", 400)
		return
	}
	user := c.MustGet("user").(*model.User)
	if !user.CanRemove() {
		common.ErrorResp(c, errs.PermissionDenied, 403)
		return
	}
	req.Dir = stdpath.Join(user.BasePath, req.Dir)
	for _, name := range req.Names {
		err := fs.Remove(c, stdpath.Join(req.Dir, name))
		if err != nil {
			common.ErrorResp(c, err, 500)
			return
		}
	}
	fs.ClearCache(req.Dir)
	common.SuccessResp(c)
}

func FsPut(c *gin.Context) {
	path := c.GetHeader("File-Path")
	user := c.MustGet("user").(*model.User)
	path = stdpath.Join(user.BasePath, path)
	if !user.CanWrite() {
		meta, err := db.GetNearestMeta(path)
		if err != nil {
			common.ErrorResp(c, err, 500)
			return
		}
		if !canWrite(meta, path) {
			common.ErrorResp(c, errs.PermissionDenied, 403)
			return
		}
	}

	dir, name := stdpath.Split(path)
	sizeStr := c.GetHeader("Content-Length")
	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	if err := fs.PutAsTask(dir, model.FileStream{
		Obj: model.Object{
			Name:     name,
			Size:     size,
			Modified: time.Now(),
		},
		ReadCloser: c.Request.Body,
		Mimetype:   c.GetHeader("Content-Type"),
	}); err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c)
}

// Link return real link, just for proxy program, it may contain cookie
func Link(c *gin.Context) {
	var req FsGetOrLinkReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	user := c.MustGet("user").(*model.User)
	rawPath := stdpath.Join(user.BasePath, req.Path)
	account, err := fs.GetAccount(rawPath)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	if account.Config().OnlyLocal {
		common.SuccessResp(c, model.Link{
			URL: fmt.Sprintf("%s/p%s?d&sign=%s", common.GetBaseUrl(c.Request), req.Path, sign.Sign(stdpath.Base(rawPath))),
		})
		return
	}
	link, _, err := fs.Link(c, rawPath, model.LinkArgs{IP: c.ClientIP()})
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c, link)
	return
}
