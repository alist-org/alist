package server

import (
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"path/filepath"
	"strings"
)

type PathReq struct {
	Path     string `json:"Path"`
	Password string `json:"Password"`
}

func Path(c *gin.Context) {
	var req PathReq
	if err := c.ShouldBind(&req); err != nil {
		ErrorResp(c, err, 400)
		return
	}
	req.Path = utils.ParsePath(req.Path)
	log.Debugf("path: %s", req.Path)
	meta, err := model.GetMetaByPath(req.Path)
	if err == nil {
		if meta.Password != "" && meta.Password != req.Password {
			ErrorResp(c, fmt.Errorf("wrong password"), 401)
			return
		}
		// TODO hide or ignore?
	} else if conf.CheckParent {
		if !CheckParent(filepath.Dir(req.Path), req.Password) {
			ErrorResp(c, fmt.Errorf("wrong password"), 401)
			return
		}
	}
	if model.AccountsCount() > 1 && req.Path == "/" {
		files, err := model.GetAccountFiles()
		if err != nil {
			ErrorResp(c, err, 500)
			return
		}
		c.JSON(200, Resp{
			Code:    200,
			Message: "folder",
			Data:    files,
		})
		return
	}
	account, path, driver, err := ParsePath(req.Path)
	if err != nil {
		ErrorResp(c, err, 500)
		return
	}
	file, files, err := driver.Path(path, account)
	if err != nil {
		ErrorResp(c, err, 500)
		return
	}
	if file != nil {
		if account.Type == "Native" {
			file.Url = fmt.Sprintf("//%s/d%s", c.Request.Host, req.Path)
		}
		c.JSON(200, Resp{
			Code:    200,
			Message: "file",
			Data:    []*model.File{file},
		})
	} else {
		if meta != nil && meta.Hide != "" {
			tmpFiles := make([]model.File, 0)
			hideFiles := strings.Split(meta.Hide, ",")
			for _, item := range files {
				if !utils.IsContain(hideFiles, item.Name) {
					tmpFiles = append(tmpFiles, item)
				}
			}
			files = tmpFiles
		}
		c.JSON(200, Resp{
			Code:    200,
			Message: "folder",
			Data:    files,
		})
	}
}

func Link(c *gin.Context) {
	var req PathReq
	if err := c.ShouldBind(&req); err != nil {
		ErrorResp(c, err, 400)
		return
	}
	rawPath := req.Path
	rawPath = utils.ParsePath(rawPath)
	log.Debugf("link: %s", rawPath)
	account, path, driver, err := ParsePath(rawPath)
	if err != nil {
		ErrorResp(c, err, 500)
		return
	}
	link, err := driver.Link(path, account)
	if err != nil {
		ErrorResp(c, err, 500)
		return
	}
	if account.Type == "Native" {
		SuccessResp(c, gin.H{
			"url": fmt.Sprintf("//%s/d%s", c.Request.Host, req.Path),
		})
		return
	} else {
		SuccessResp(c, gin.H{
			"url": link,
		})
		return
	}
}

func Preview(c *gin.Context) {
	var req PathReq
	if err := c.ShouldBind(&req); err != nil {
		ErrorResp(c, err, 400)
		return
	}
	rawPath := req.Path
	rawPath = utils.ParsePath(rawPath)
	log.Debugf("preview: %s", rawPath)
	account, path, driver, err := ParsePath(rawPath)
	if err != nil {
		ErrorResp(c, err, 500)
		return
	}
	data, err := driver.Preview(path, account)
	if err != nil {
		ErrorResp(c, err, 500)
	} else {
		SuccessResp(c, data)
	}
}
