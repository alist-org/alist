package alist

import (
	"errors"
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"path/filepath"
	"strings"
	"time"
)

type Alist struct{}

func (driver Alist) Config() base.DriverConfig {
	return base.DriverConfig{
		Name:          "Alist",
		NoNeedSetLink: true,
		NoCors:        true,
	}
}

func (driver Alist) Items() []base.Item {
	return []base.Item{
		{
			Name:     "site_url",
			Label:    "alist site url",
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
	if account == nil {
		return nil
	}
	account.SiteUrl = strings.TrimRight(account.SiteUrl, "/")
	if account.RootFolder == "" {
		account.RootFolder = "/"
	}
	err := driver.Login(account)
	if err == nil {
		account.Status = "work"
	} else {
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
	_, files, err := driver.Path(path, account)
	if err != nil {
		return nil, err
	}
	if files == nil {
		return nil, base.ErrNotFolder
	}
	return files, nil
}

func (driver Alist) Link(args base.Args, account *model.Account) (*base.Link, error) {
	path := args.Path
	path = utils.ParsePath(path)
	name := utils.Base(path)
	flag := "d"
	if utils.GetFileType(filepath.Ext(path)) == conf.TEXT {
		flag = "p"
	}
	link := base.Link{}
	link.Url = fmt.Sprintf("%s/%s%s?sign=%s", account.SiteUrl, flag, utils.Join(utils.ParsePath(account.RootFolder), path), utils.SignWithToken(name, conf.Token))
	return &link, nil
}

func (driver Alist) Path(path string, account *model.Account) (*model.File, []model.File, error) {
	path = utils.ParsePath(path)
	path = filepath.Join(account.RootFolder, path)
	path = strings.ReplaceAll(path, "\\", "/")
	cache, err := base.GetCache(path, account)
	if err == nil {
		files := cache.([]model.File)
		return nil, files, nil
	}
	var resp PathResp
	_, err = base.RestyClient.R().SetResult(&resp).
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
	if resp.Data.Type == "file" {
		return &resp.Data.Files[0], nil, nil
	}
	if len(resp.Data.Files) > 0 {
		_ = base.SetCache(path, resp.Data.Files, account)
	}
	return nil, resp.Data.Files, nil
}

//func (driver Alist) Proxy(r *http.Request, account *model.Account) {}

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

func (driver Alist) Rename(src string, dst string, account *model.Account) error {
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
