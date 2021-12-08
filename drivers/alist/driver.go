package alist

import (
	"errors"
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/gin-gonic/gin"
	"path/filepath"
	"strings"
	"time"
)

type Alist struct{}

func (driver Alist) Config() base.DriverConfig {
	return base.DriverConfig{
		Name:      "Alist",
		OnlyProxy: false,
	}
}

func (driver Alist) Items() []base.Item {
	return []base.Item{
		{
			Name:     "site_url",
			Label:    "site url",
			Type:     base.TypeString,
			Required: true,
		},
		{
			Name:        "access_token",
			Label:       "token",
			Type:        base.TypeString,
			Description: "admin token",
			Required:    true,
		},
		{
			Name:     "root_folder",
			Label:    "root folder path",
			Type:     base.TypeString,
			Required: false,
		},
	}
}

func (driver Alist) Save(account *model.Account, old *model.Account) error {
	account.SiteUrl = strings.TrimRight(account.SiteUrl, "/")
	if account.RootFolder == "" {
		account.RootFolder = "/"
	}
	err := driver.Login(account)
	if err == nil {
		account.Status = "work"
	}else {
		account.Status = err.Error()
	}
	_ = model.SaveAccount(account)
	return err
}

func (driver Alist) File(path string, account *model.Account) (*model.File, error) {
	now := time.Now()
	if path == "/" {
		return &model.File{
			Id:        "root",
			Name:      "root",
			Size:      0,
			Type:      conf.FOLDER,
			Driver:    driver.Config().Name,
			UpdatedAt: &now,
		}, nil
	}
	_, files, err := driver.Path(utils.Dir(path), account)
	if err != nil {
		return nil, err
	}
	if files == nil {
		return nil, base.ErrPathNotFound
	}
	name := utils.Base(path)
	for _, file := range files {
		if file.Name == name {
			return &file, nil
		}
	}
	return nil, base.ErrPathNotFound
}

func (driver Alist) Files(path string, account *model.Account) ([]model.File, error) {
	//return nil, base.ErrNotImplement
	return []model.File{}, nil
}

func (driver Alist) Link(path string, account *model.Account) (string, error) {
	path = utils.ParsePath(path)
	name := utils.Base(path)
	flag := "d"
	if utils.GetFileType(filepath.Ext(path)) == conf.TEXT {
		flag = "p"
	}
	return fmt.Sprintf("%s/%s%s?sign=%s", account.SiteUrl, flag, path, utils.SignWithToken(name,conf.Token)), nil
}

func (driver Alist) Path(path string, account *model.Account) (*model.File, []model.File, error) {
	path = utils.ParsePath(path)
	var resp PathResp
	_, err := base.RestyClient.R().SetResult(&resp).
		SetHeader("Authorization", account.AccessToken).
		SetBody(base.Json{
			"path": path,
		}).Post(account.SiteUrl + "/api/public/path")
	if err != nil {
		return nil, nil, err
	}
	if resp.Code != 200 {
		return nil, nil, errors.New(resp.Message)
	}
	if resp.Message == "file" {
		return &resp.Data[0], nil, nil
	}
	_ = base.SetCache(path, resp.Data, account)
	return nil, resp.Data, nil
}

func (driver Alist) Proxy(c *gin.Context, account *model.Account) {}

func (driver Alist) Preview(path string, account *model.Account) (interface{}, error) {
	var resp PathResp
	_, err := base.RestyClient.R().SetResult(&resp).
		SetHeader("Authorization", account.AccessToken).
		SetBody(base.Json{
			"path": path,
		}).Post(account.SiteUrl + "/api/public/preview")
	if err != nil {
		return nil, err
	}
	if resp.Code != 200 {
		return nil, errors.New(resp.Message)
	}
	return resp.Data, nil
}

func (driver Alist) MakeDir(path string, account *model.Account) error {
	return base.ErrNotImplement
}

func (driver Alist) Move(src string, dst string, account *model.Account) error {
	return base.ErrNotImplement
}

func (driver Alist) Copy(src string, dst string, account *model.Account) error {
	return base.ErrNotImplement
}

func (driver Alist) Delete(path string, account *model.Account) error {
	return base.ErrNotImplement
}

func (driver Alist) Upload(file *model.FileStream, account *model.Account) error {
	return base.ErrNotImplement
}

var _ base.Driver = (*Alist)(nil)
