package controllers

import (
	"errors"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/server/common"
	"github.com/Xhofe/alist/utils"
	"github.com/gin-gonic/gin"
)

func UploadFile(c *gin.Context) {
	path := c.PostForm("path")
	path = utils.ParsePath(path)
	token := c.GetHeader("Authorization")
	if token != conf.Token {
		password := c.PostForm("password")
		meta, _ := model.GetMetaByPath(path)
		if meta == nil || !meta.Upload {
			common.ErrorResp(c, errors.New("not allow upload"), 403)
			return
		}
		if meta.Password != "" && meta.Password != password {
			common.ErrorResp(c, errors.New("wrong password"), 403)
			return
		}
	}
	file, err := c.FormFile("file")
	if err != nil {
		common.ErrorResp(c, err, 400)
	}
	open, err := file.Open()
	defer func() {
		_ = open.Close()
	}()
	if err != nil {
		return
	}
	account, path_, driver, err := common.ParsePath(path)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	fileStream := model.FileStream{
		File:       open,
		Size:       uint64(file.Size),
		ParentPath: path_,
		Name:       file.Filename,
		MIMEType:   file.Header.Get("Content-Type"),
	}
	err = driver.Upload(&fileStream, account)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c)
}
