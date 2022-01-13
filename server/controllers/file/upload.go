package file

import (
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/drivers/operate"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/server/common"
	"github.com/Xhofe/alist/utils"
	"github.com/gin-gonic/gin"
)

func UploadFiles(c *gin.Context) {
	path := c.PostForm("path")
	path = utils.ParsePath(path)
	token := c.GetHeader("Authorization")
	if token != conf.Token {
		password := c.PostForm("password")
		meta, _ := model.GetMetaByPath(path)
		if meta == nil || !meta.Upload {
			common.ErrorStrResp(c, "Not allow upload", 403)
			return
		}
		if meta.Password != "" && meta.Password != password {
			common.ErrorStrResp(c, "Wrong password", 403)
			return
		}
	}
	account, path_, driver, err := common.ParsePath(path)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	form, err := c.MultipartForm()
	if err != nil {
		common.ErrorResp(c, err, 400)
	}
	files := form.File["files"]
	if err != nil {
		return
	}
	for i, file := range files {
		open, err := file.Open()
		fileStream := model.FileStream{
			File:       open,
			Size:       uint64(file.Size),
			ParentPath: path_,
			Name:       file.Filename,
			MIMEType:   file.Header.Get("Content-Type"),
		}
		clearCache := false
		if i == len(files)-1 {
			clearCache = true
		}
		err = operate.Upload(driver, account, &fileStream, clearCache)
		if err != nil {
			if i != 0 {
				_ = base.DeleteCache(path_, account)
			}
			common.ErrorResp(c, err, 500)
			return
		}
	}
	common.SuccessResp(c)
}
