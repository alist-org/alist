package handles

import (
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/pkg/utils/random"
	"io"
	"net/url"
	"os"
	stdpath "path"
	"reflect"
	"strconv"
	"time"

	"github.com/alist-org/alist/v3/internal/stream"

	"github.com/alist-org/alist/v3/internal/fs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/gin-gonic/gin"
)

func getLastModified(c *gin.Context) time.Time {
	now := time.Now()
	lastModifiedStr := c.GetHeader("Last-Modified")
	lastModifiedMillisecond, err := strconv.ParseInt(lastModifiedStr, 10, 64)
	if err != nil {
		return now
	}
	lastModified := time.UnixMilli(lastModifiedMillisecond)
	return lastModified
}

func FsStream(c *gin.Context) {
	path := c.GetHeader("File-Path")
	path, err := url.PathUnescape(path)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	asTask := c.GetHeader("As-Task") == "true"
	user := c.MustGet("user").(*model.User)
	path, err = user.JoinPath(path)
	if err != nil {
		common.ErrorResp(c, err, 403)
		return
	}
	dir, name := stdpath.Split(path)
	sizeStr := c.GetHeader("Content-Length")
	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	s := &stream.FileStream{
		Obj: &model.Object{
			Name:     name,
			Size:     size,
			Modified: getLastModified(c),
		},
		Reader:       c.Request.Body,
		Mimetype:     c.GetHeader("Content-Type"),
		WebPutAsTask: asTask,
	}
	if asTask {
		err = fs.PutAsTask(dir, s)
	} else {
		err = fs.PutDirectly(c, dir, s, true)
	}
	defer c.Request.Body.Close()
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
	path, err = user.JoinPath(path)
	if err != nil {
		common.ErrorResp(c, err, 403)
		return
	}
	storage, err := fs.GetStorage(path, &fs.GetStoragesArgs{})
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
	tmpFile, tmpInSys := "", ""
	fv := reflect.ValueOf(*file)
	tmpInSys = fv.FieldByName("tmpfile").String()

	var f io.Reader
	var osFile *os.File
	if len(tmpInSys) > 0 {
		tmpFile = conf.Conf.TempDir + "file-" + random.String(8)
		err = os.Rename(tmpInSys, tmpFile)
		if err != nil {
			common.ErrorResp(c, err, 500)
			return
		}
		osFile, err = os.Open(tmpFile)
		f = osFile
	} else {
		f, err = file.Open()
	}
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	dir, name := stdpath.Split(path)
	s := stream.FileStream{
		Obj: &model.Object{
			Name:     name,
			Size:     file.Size,
			Modified: getLastModified(c),
		},
		Reader:       f,
		Mimetype:     file.Header.Get("Content-Type"),
		WebPutAsTask: asTask,
	}
	ss, err := stream.NewSeekableStream(s, nil)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	if osFile != nil {
		ss.SetTmpFile(osFile)
	}

	if asTask {
		err = fs.PutAsTask(dir, ss)
	} else {
		err = fs.PutDirectly(c, dir, ss, true)
	}

	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c)
}
