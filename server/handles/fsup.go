package handles

import (
	"net/url"
	stdpath "path"
	"strconv"
	"time"

	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/fs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func FsStream(c *gin.Context) {
	path := c.GetHeader("File-Path")
	path, err := url.PathUnescape(path)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	asTask := c.GetHeader("As-Task") == "true"
	user := c.MustGet("user").(*model.User)
	path = stdpath.Join(user.BasePath, path)
	if !user.CanWrite() {
		meta, err := db.GetNearestMeta(stdpath.Dir(path))
		if err != nil {
			if !errors.Is(errors.Cause(err), errs.MetaNotFound) {
				common.ErrorResp(c, err, 500, true)
				return
			}
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
	stream := &model.FileStream{
		Obj: &model.Object{
			Name:     name,
			Size:     size,
			Modified: time.Now(),
		},
		ReadCloser:   c.Request.Body,
		Mimetype:     c.GetHeader("Content-Type"),
		WebPutAsTask: asTask,
	}
	if asTask {
		err = fs.PutAsTask(dir, stream)
	} else {
		err = fs.PutDirectly(c, dir, stream)
	}
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c)
}

func FsForm(c *gin.Context) {
	path := c.GetHeader("File-Path")
	path, err := url.PathUnescape(path)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	asTask := c.GetHeader("As-Task") == "true"
	user := c.MustGet("user").(*model.User)
	path = stdpath.Join(user.BasePath, path)
	if !user.CanWrite() {
		meta, err := db.GetNearestMeta(stdpath.Dir(path))
		if err != nil {
			if !errors.Is(errors.Cause(err), errs.MetaNotFound) {
				common.ErrorResp(c, err, 500, true)
				return
			}
		}
		if !canWrite(meta, path) {
			common.ErrorResp(c, errs.PermissionDenied, 403)
			return
		}
	}
	storage, err := fs.GetStorage(path)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	if storage.Config().NoUpload {
		common.ErrorStrResp(c, "Current storage doesn't support upload", 405)
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	f, err := file.Open()
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	dir, name := stdpath.Split(path)
	stream := &model.FileStream{
		Obj: &model.Object{
			Name:     name,
			Size:     file.Size,
			Modified: time.Now(),
		},
		ReadCloser:   f,
		Mimetype:     file.Header.Get("Content-Type"),
		WebPutAsTask: false,
	}
	if asTask {
		err = fs.PutAsTask(dir, stream)
	} else {
		err = fs.PutDirectly(c, dir, stream)
	}
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c)
}
