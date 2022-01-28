package controllers

import (
	"errors"
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/drivers/operate"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/server/common"
	"github.com/Xhofe/alist/utils"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func Pagination(files []model.File, req *common.PathReq) (int, []model.File) {
	pageNum, pageSize := req.PageNum, req.PageSize
	total := len(files)
	if isAll(req) {
		return total, files
	}
	switch conf.GetStr("load type") {
	case "all":
		return total, files
		//case "pagination":
		//
	}
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

func isAll(req *common.PathReq) bool {
	return req.PageNum == 0 && req.PageSize == 0
}

func CheckPagination(req *common.PathReq) error {
	if isAll(req) {
		return nil
	}
	if conf.GetStr("loading type") == "all" {
		return nil
	}
	if req.PageNum < 1 {
		return errors.New("page_num can't be less than 1")
	}
	if req.PageSize == 0 {
		req.PageSize = conf.GetInt("default page size", 30)
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
	_, ok := c.Get("admin")
	meta, _ := model.GetMetaByPath(req.Path)
	upload := false
	if meta != nil && meta.Upload {
		upload = true
	}
	if model.AccountsCount() > 1 && (req.Path == "/" || req.Path == "") {
		files, err := model.GetAccountFiles()
		if err != nil {
			common.ErrorResp(c, err, 500)
			return
		}
		if !ok {
			files = common.Hide(meta, files)
		}
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
	err := CheckPagination(&req)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	account, path, driver, err := common.ParsePath(req.Path)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	file, files, err := operate.Path(driver, account, path)
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
		if !ok {
			files = common.Hide(meta, files)
		}
		if driver.Config().LocalSort {
			model.SortFiles(files, account)
		}
		model.ExtractFolder(files, account)
		total, files := Pagination(files, &req)
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
