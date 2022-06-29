package controllers

import (
	"fmt"
	"github.com/alist-org/alist/v3/internal/fs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/gin-gonic/gin"
	stdpath "path"
	"strconv"
	"time"
)

type MkdirReq struct {
	Path string `json:"path"`
}

func FsMkdir(c *gin.Context) {
	var req MkdirReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	user := c.MustGet("user").(*model.User)
	req.Path = stdpath.Join(user.BasePath, req.Path)
	if err := fs.MakeDir(c, req.Path); err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c)
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
	req.SrcDir = stdpath.Join(user.BasePath, req.SrcDir)
	req.DstDir = stdpath.Join(user.BasePath, req.DstDir)
	for _, name := range req.Names {
		err := fs.Move(c, stdpath.Join(req.SrcDir, name), req.DstDir)
		if err != nil {
			common.ErrorResp(c, err, 500)
			return
		}
	}
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
	req.Path = stdpath.Join(user.BasePath, req.Path)
	if err := fs.Rename(c, req.Path, req.Name); err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c)
}

type RemoveReq struct {
	Path  string   `json:"path"`
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
	req.Path = stdpath.Join(user.BasePath, req.Path)
	for _, name := range req.Names {
		err := fs.Remove(c, stdpath.Join(req.Path, name))
		if err != nil {
			common.ErrorResp(c, err, 500)
			return
		}
	}
	common.SuccessResp(c)
}

func FsPut(c *gin.Context) {
	path := c.GetHeader("File-Path")
	user := c.MustGet("user").(*model.User)
	path = stdpath.Join(user.BasePath, path)
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
