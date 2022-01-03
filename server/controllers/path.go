package controllers

import (
	"errors"
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/server/common"
	"github.com/Xhofe/alist/utils"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"strings"
)

func Hide(meta *model.Meta, files []model.File) []model.File {
	//meta, _ := model.GetMetaByPath(path)
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
	return files
}

func Pagination(files []model.File, pageNum, pageSize int) (int, []model.File) {
	total := len(files)
	start := (pageNum - 1) * pageSize
	if start > total {
		return total, []model.File{}
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return total, files[start:end]
}

func CheckPagination(req common.PathReq) error {
	if req.PageNum < 1 {
		return errors.New("page_num can't be less than 1")
	}
	return nil
}

type Meta struct {
	Driver string `json:"driver"`
	Upload bool   `json:"upload"`
	Total  int    `json:"total"`
	//Pages  int    `json:"pages"`
}

type PathResp struct {
	Type  string       `json:"type"`
	Meta  Meta         `json:"meta"`
	Files []model.File `json:"files"`
}

func Path(c *gin.Context) {
	reqV, _ := c.Get("req")
	req := reqV.(common.PathReq)
	meta, _ := model.GetMetaByPath(req.Path)
	upload := false
	if meta != nil && meta.Upload {
		upload = true
	}
	if model.AccountsCount() > 1 && req.Path == "/" {
		files, err := model.GetAccountFiles()
		if err != nil {
			common.ErrorResp(c, err, 500)
			return
		}
		files = Hide(meta, files)
		c.JSON(200, common.Resp{
			Code:    200,
			Message: "success",
			Data: PathResp{
				Type: "folder",
				Meta: Meta{
					Driver: "root",
				},
				Files: files,
			},
		})
		return
	}
	err := CheckPagination(req)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	account, path, driver, err := common.ParsePath(req.Path)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	file, files, err := driver.Path(path, account)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	if file != nil {
		// 对于中转文件或只能中转,将链接修改为中转链接
		if driver.Config().OnlyProxy || account.Proxy {
			if account.DownProxyUrl != "" {
				file.Url = fmt.Sprintf("%s%s?sign=%s", account.DownProxyUrl, req.Path, utils.SignWithToken(file.Name, conf.Token))
			} else {
				file.Url = fmt.Sprintf("//%s/p%s?sign=%s", c.Request.Host, req.Path, utils.SignWithToken(file.Name, conf.Token))
			}
		} else if !driver.Config().NoNeedSetLink {
			link, err := driver.Link(base.Args{Path: path, IP: c.ClientIP()}, account)
			if err != nil {
				common.ErrorResp(c, err, 500)
				return
			}
			file.Url = link.Url
		}
		c.JSON(200, common.Resp{
			Code:    200,
			Message: "success",
			Data: PathResp{
				Type: "file",
				Meta: Meta{
					Driver: driver.Config().Name,
				},
				Files: []model.File{*file},
			},
		})
	} else {
		files = Hide(meta, files)
		if driver.Config().LocalSort {
			model.SortFiles(files, account)
		}
		total, files := Pagination(files, req.PageNum, req.PageSize)
		c.JSON(200, common.Resp{
			Code:    200,
			Message: "success",
			Data: PathResp{
				Type: "folder",
				Meta: Meta{
					Driver: driver.Config().Name,
					Upload: upload,
					Total:  total,
				},
				Files: files,
			},
		})
	}
}

func Preview(c *gin.Context) {
	reqV, _ := c.Get("req")
	req := reqV.(common.PathReq)
	rawPath := req.Path
	rawPath = utils.ParsePath(rawPath)
	log.Debugf("preview: %s", rawPath)
	account, path, driver, err := common.ParsePath(rawPath)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	data, err := driver.Preview(path, account)
	if err != nil {
		common.ErrorResp(c, err, 500)
	} else {
		common.SuccessResp(c, data)
	}
}
