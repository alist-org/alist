package uss

import (
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	log "github.com/sirupsen/logrus"
	"github.com/upyun/go-sdk/v3/upyun"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

type USS struct {
}

func (driver USS) Config() base.DriverConfig {
	return base.DriverConfig{
		Name:      "USS",
		LocalSort: true,
	}
}

func (driver USS) Items() []base.Item {
	return []base.Item{
		{
			Name:     "bucket",
			Label:    "Bucket",
			Type:     base.TypeString,
			Required: true,
		},
		{
			Name:     "endpoint",
			Label:    "Endpoint",
			Type:     base.TypeString,
			Required: true,
		},
		{
			Name:     "access_key",
			Label:    "Operator Name",
			Type:     base.TypeString,
			Required: true,
		},
		{
			Name:     "access_secret",
			Label:    "Operator Password",
			Type:     base.TypeString,
			Required: true,
		},
		{
			Name:     "root_folder",
			Label:    "root folder path",
			Type:     base.TypeString,
			Required: false,
		},
		{
			Name:  "custom_host",
			Label: "Custom Host",
			Type:  base.TypeString,
		},
		{
			Name:        "limit",
			Label:       "Sign url expire time(hours)",
			Type:        base.TypeNumber,
			Default:     "4",
			Description: "default 4 hours",
		},
		//{
		//	Name:        "zone",
		//	Label:       "placeholder filename",
		//	Type:        base.TypeString,
		//	Description: "default empty string",
		//},
	}
}

func (driver USS) Save(account *model.Account, old *model.Account) error {
	if account == nil {
		return nil
	}
	if account.Limit == 0 {
		account.Limit = 4
	}
	client, err := driver.NewUpYun(account)
	if err != nil {
		account.Status = err.Error()
	} else {
		clientsMap[account.Name] = client
		account.Status = "work"
	}
	_ = model.SaveAccount(account)
	return err
}

func (driver USS) File(path string, account *model.Account) (*model.File, error) {
	path = utils.ParsePath(path)
	if path == "/" {
		return &model.File{
			Id:        account.RootFolder,
			Name:      account.Name,
			Size:      0,
			Type:      conf.FOLDER,
			Driver:    driver.Config().Name,
			UpdatedAt: account.UpdatedAt,
		}, nil
	}
	dir, name := filepath.Split(path)
	files, err := driver.Files(dir, account)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if file.Name == name {
			return &file, nil
		}
	}
	return nil, base.ErrPathNotFound
}

func (driver USS) Files(path string, account *model.Account) ([]model.File, error) {
	path = utils.ParsePath(path)
	var files []model.File
	cache, err := base.GetCache(path, account)
	if err == nil {
		files, _ = cache.([]model.File)
	} else {
		files, err = driver.List(path, account)
		if err == nil && len(files) > 0 {
			_ = base.SetCache(path, files, account)
		}
	}
	return files, err
}

func (driver USS) Link(args base.Args, account *model.Account) (*base.Link, error) {
	key := driver.GetKey(args.Path, account, false)
	host := account.CustomHost
	if host == "" {
		host = account.Endpoint
	}
	if strings.Contains(host, "://") {
		host = "https://" + host
	}
	u := fmt.Sprintf("%s/%s", host, key)
	downExp := time.Hour * time.Duration(account.Limit)
	expireAt := time.Now().Add(downExp).Unix()
	upd := url.QueryEscape(utils.Base(args.Path))
	signStr := strings.Join([]string{account.AccessSecret, fmt.Sprint(expireAt), fmt.Sprintf("/%s", key)}, "&")
	upt := utils.GetMD5Encode(signStr)[12:20] + fmt.Sprint(expireAt)
	link := fmt.Sprintf("%s?_upd=%s&_upt=%s", u, upd, upt)
	return &base.Link{Url: link}, nil
}

func (driver USS) Path(path string, account *model.Account) (*model.File, []model.File, error) {
	path = utils.ParsePath(path)
	log.Debugf("s3 path: %s", path)
	file, err := driver.File(path, account)
	if err != nil {
		return nil, nil, err
	}
	if !file.IsDir() {
		return file, nil, nil
	}
	files, err := driver.Files(path, account)
	if err != nil {
		return nil, nil, err
	}
	return nil, files, nil
}

func (driver USS) Preview(path string, account *model.Account) (interface{}, error) {
	return nil, base.ErrNotSupport
}

func (driver USS) MakeDir(path string, account *model.Account) error {
	client, err := driver.GetClient(account)
	if err != nil {
		return err
	}
	return client.Mkdir(driver.GetKey(path, account, true))
}

func (driver USS) Move(src string, dst string, account *model.Account) error {
	client, err := driver.GetClient(account)
	if err != nil {
		return err
	}
	return client.Move(&upyun.MoveObjectConfig{
		SrcPath:  driver.GetKey(src, account, false),
		DestPath: driver.GetKey(dst, account, false),
	})
}

func (driver USS) Rename(src string, dst string, account *model.Account) error {
	return driver.Move(src, dst, account)
}

func (driver USS) Copy(src string, dst string, account *model.Account) error {
	client, err := driver.GetClient(account)
	if err != nil {
		return err
	}
	return client.Copy(&upyun.CopyObjectConfig{
		SrcPath:  driver.GetKey(src, account, false),
		DestPath: driver.GetKey(dst, account, false),
	})
}

func (driver USS) Delete(path string, account *model.Account) error {
	client, err := driver.GetClient(account)
	if err != nil {
		return err
	}
	return client.Delete(&upyun.DeleteObjectConfig{
		Path:  driver.GetKey(path, account, false),
		Async: false,
	})
}

func (driver USS) Upload(file *model.FileStream, account *model.Account) error {
	if file == nil {
		return base.ErrEmptyFile
	}
	client, err := driver.GetClient(account)
	if err != nil {
		return err
	}
	return client.Put(&upyun.PutObjectConfig{
		Path:   driver.GetKey(utils.Join(file.ParentPath, file.GetFileName()), account, false),
		Reader: file,
	})
}

var _ base.Driver = (*USS)(nil)
