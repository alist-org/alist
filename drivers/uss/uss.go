package uss

import (
	"errors"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/upyun/go-sdk/v3/upyun"
	"path"
	"strings"
)

var clientsMap map[string]*upyun.UpYun

func (driver USS) NewUpYun(account *model.Account) (*upyun.UpYun, error) {
	return upyun.NewUpYun(&upyun.UpYunConfig{
		Bucket:   account.Bucket,
		Operator: account.AccessKey,
		Password: account.AccessToken,
	}), nil
}

func (driver USS) GetClient(account *model.Account) (*upyun.UpYun, error) {
	client, ok := clientsMap[account.Name]
	if ok {
		return client, nil
	}
	return nil, errors.New("can't get client")
}

func (driver USS) List(prefix string, account *model.Account) ([]model.File, error) {
	prefix = driver.GetKey(prefix, account, true)
	client, err := driver.GetClient(account)
	if err != nil {
		return nil, err
	}
	objsChan := make(chan *upyun.FileInfo, 10)
	defer close(objsChan)
	go func() {
		err = client.List(&upyun.GetObjectsConfig{
			Path:           prefix,
			ObjectsChan:    objsChan,
			MaxListObjects: 0,
			MaxListLevel:   1,
		})
	}()
	if err != nil {
		return nil, err
	}
	res := make([]model.File, 0)
	for obj := range objsChan {
		t := obj.Time
		f := model.File{
			Name:      obj.Name,
			Size:      obj.Size,
			UpdatedAt: &t,
			Driver:    driver.Config().Name,
		}
		if obj.IsDir {
			f.Type = conf.FOLDER
		} else {
			f.Type = utils.GetFileType(path.Ext(obj.Name))
		}
		res = append(res, f)
	}
	return res, err
}

func (driver USS) GetKey(path string, account *model.Account, dir bool) string {
	path = utils.Join(account.RootFolder, path)
	path = strings.TrimPrefix(path, "/")
	if dir {
		path += "/"
	}
	return path
}

func init() {
	clientsMap = make(map[string]*upyun.UpYun)
	base.RegisterDriver(&USS{})
}
