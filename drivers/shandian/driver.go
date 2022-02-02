package shandian

import (
	"errors"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	log "github.com/sirupsen/logrus"
	"path/filepath"
	"strconv"
)

type Shandian struct{}

func (driver Shandian) Config() base.DriverConfig {
	return base.DriverConfig{
		Name:          "ShandianPan",
		NoNeedSetLink: true,
		LocalSort:     true,
	}
}

func (driver Shandian) Items() []base.Item {
	return []base.Item{
		{
			Name:        "username",
			Label:       "username",
			Type:        base.TypeString,
			Required:    true,
			Description: "account username/phone number",
		},
		{
			Name:        "password",
			Label:       "password",
			Type:        base.TypeString,
			Required:    true,
			Description: "account password",
		},
		{
			Name:     "root_folder",
			Label:    "root folder file_id",
			Type:     base.TypeString,
			Required: false,
		},
	}
}

func (driver Shandian) Save(account *model.Account, old *model.Account) error {
	if account == nil {
		return nil
	}
	if account.RootFolder == "" {
		account.RootFolder = "0"
	}
	return driver.Login(account)
}

func (driver Shandian) File(path string, account *model.Account) (*model.File, error) {
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

func (driver Shandian) Files(path string, account *model.Account) ([]model.File, error) {
	path = utils.ParsePath(path)
	var files []model.File
	cache, err := base.GetCache(path, account)
	if err == nil {
		files, _ = cache.([]model.File)
	} else {
		file, err := driver.File(path, account)
		if err != nil {
			return nil, err
		}
		rawFiles, err := driver.GetFiles(file.Id, account)
		if err != nil {
			return nil, err
		}
		files = make([]model.File, 0)
		for _, file := range rawFiles {
			files = append(files, *driver.FormatFile(&file))
		}
		if len(files) > 0 {
			_ = base.SetCache(path, files, account)
		}
	}
	return files, nil
}

func (driver Shandian) Link(args base.Args, account *model.Account) (*base.Link, error) {
	log.Debugf("shandian link")
	file, err := driver.File(args.Path, account)
	if err != nil {
		return nil, err
	}
	var e Resp
	res, err := base.NoRedirectClient.R().SetError(&e).SetHeader("Accept", "application/json").SetQueryParams(map[string]string{
		"id":    file.Id,
		"token": account.AccessToken,
	}).Get("https://shandianpan.com/api/pan/file-download")
	if err != nil {
		return nil, err
	}
	if e.Code != 0 {
		if e.Code == 10 {
			err = driver.Login(account)
			if err != nil {
				return nil, err
			}
			return driver.Link(args, account)
		}
		return nil, errors.New(e.Msg)
	}
	return &base.Link{
		Url: res.Header().Get("location"),
	}, nil
}

func (driver Shandian) Path(path string, account *model.Account) (*model.File, []model.File, error) {
	path = utils.ParsePath(path)
	log.Debugf("shandian path: %s", path)
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

//func (driver Shandian) Proxy(r *http.Request, account *model.Account) {
//
//}

func (driver Shandian) Preview(path string, account *model.Account) (interface{}, error) {
	return nil, base.ErrNotSupport
}

func (driver Shandian) MakeDir(path string, account *model.Account) error {
	parentFile, err := driver.File(utils.Dir(path), account)
	if err != nil {
		return err
	}
	data := map[string]interface{}{
		"id":   parentFile.Id,
		"name": utils.Base(path),
	}
	_, err = driver.Post("https://shandianpan.com/api/pan/mkdir", data, nil, account)
	return err
}

func (driver Shandian) Move(src string, dst string, account *model.Account) error {
	srcFile, err := driver.File(src, account)
	if err != nil {
		return err
	}
	dstParentFile, err := driver.File(utils.Dir(dst), account)
	if err != nil {
		return err
	}
	data := map[string]interface{}{
		"id":    srcFile.Id,
		"to_id": dstParentFile.Id,
	}
	_, err = driver.Post("https://shandianpan.com/api/pan/move", data, nil, account)
	return err
}

func (driver Shandian) Rename(src string, dst string, account *model.Account) error {
	srcFile, err := driver.File(src, account)
	if err != nil {
		return err
	}
	data := map[string]interface{}{
		"id":   srcFile.Id,
		"name": utils.Base(dst),
	}
	_, err = driver.Post("https://shandianpan.com/api/pan/change", data, nil, account)
	return err
}

func (driver Shandian) Copy(src string, dst string, account *model.Account) error {
	return base.ErrNotSupport
}

func (driver Shandian) Delete(path string, account *model.Account) error {
	file, err := driver.File(path, account)
	if err != nil {
		return err
	}
	data := map[string]interface{}{
		"id": file.Id,
	}
	_, err = driver.Post("https://shandianpan.com/api/pan/recycle-in", data, nil, account)
	return err
}

func (driver Shandian) Upload(file *model.FileStream, account *model.Account) error {
	if file == nil {
		return base.ErrEmptyFile
	}
	parentFile, err := driver.File(file.ParentPath, account)
	if err != nil {
		return err
	}
	var resp UploadResp
	parentId, err := strconv.Atoi(parentFile.Id)
	if err != nil {
		return err
	}
	data := map[string]interface{}{
		"id":   parentId,
		"name": file.GetFileName(),
	}
	res, err := driver.Post("https://shandianpan.com/api/pan/upload", data, nil, account)
	if err != nil {
		return err
	}
	err = utils.Json.Unmarshal(res, &resp)
	if err != nil {
		return err
	}
	if resp.Code != 0 {
		if resp.Code == 10 {
			err = driver.Login(account)
			if err != nil {
				return err
			}
			return driver.Upload(file, account)
		}
		return errors.New(resp.Msg)
	}
	var r Resp
	_, err = base.RestyClient.R().SetMultipartFormData(map[string]string{
		"token":          account.AccessToken,
		"id":             "0",
		"key":            resp.Data.Key,
		"ossAccessKeyId": resp.Data.Accessid,
		"policy":         resp.Data.Policy,
		"signature":      resp.Data.Signature,
		"callback":       resp.Data.Callback,
	}).SetMultipartField("file", file.GetFileName(), file.GetMIMEType(), file).
		SetResult(&r).SetError(&r).Post("https:" + resp.Data.Host + "/")
	if err != nil {
		return err
	}
	if r.Code == 0 {
		return nil
	}
	return errors.New(r.Msg)
}

var _ base.Driver = (*Shandian)(nil)
