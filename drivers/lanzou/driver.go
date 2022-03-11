package lanzou

import (
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	log "github.com/sirupsen/logrus"
	"path/filepath"
)

type Lanzou struct{}

func (driver Lanzou) Config() base.DriverConfig {
	return base.DriverConfig{
		Name:   "Lanzou",
		NoCors: true,
	}
}

func (driver Lanzou) Items() []base.Item {
	return []base.Item{
		{
			Name:     "internal_type",
			Label:    "lanzou type",
			Type:     base.TypeSelect,
			Required: true,
			Values:   "cookie,url",
		},
		{
			Name:        "access_token",
			Label:       "cookie",
			Type:        base.TypeString,
			Description: "about 15 days valid",
			Required:    true,
		},
		{
			Name:     "site_url",
			Label:    "share url",
			Type:     base.TypeString,
			Required: true,
		},
		{
			Name:  "root_folder",
			Label: "root folder file_id",
			Type:  base.TypeString,
		},
		{
			Name:  "password",
			Label: "share password",
			Type:  base.TypeString,
		},
	}
}

func (driver Lanzou) Save(account *model.Account, old *model.Account) error {
	if account == nil {
		return nil
	}
	if account.InternalType == "cookie" {
		if account.RootFolder == "" {
			account.RootFolder = "-1"
		}
	}
	account.Status = "work"
	_ = model.SaveAccount(account)
	return nil
}

func (driver Lanzou) File(path string, account *model.Account) (*model.File, error) {
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

func (driver Lanzou) Files(path string, account *model.Account) ([]model.File, error) {
	path = utils.ParsePath(path)
	var rawFiles []LanZouFile
	cache, err := base.GetCache(path, account)
	if err == nil {
		rawFiles, _ = cache.([]LanZouFile)
	} else {
		file, err := driver.File(path, account)
		if err != nil {
			return nil, err
		}
		rawFiles, err = driver.GetFiles(file.Id, account)
		if err != nil {
			return nil, err
		}
		if len(rawFiles) > 0 {
			_ = base.SetCache(path, rawFiles, account)
		}
	}
	files := make([]model.File, 0)
	for _, file := range rawFiles {
		files = append(files, *driver.FormatFile(&file))
	}
	return files, nil
}

func (driver Lanzou) Link(args base.Args, account *model.Account) (*base.Link, error) {
	file, err := driver.File(args.Path, account)
	if err != nil {
		return nil, err
	}
	log.Debugf("down file: %+v", file)
	downId := file.Id
	pwd := ""
	if account.InternalType == "cookie" {
		downId, pwd, err = driver.GetDownPageId(file.Id, account)
		if err != nil {
			return nil, err
		}
	}
	var url string
	//if pwd != "" {
	//url, err = driver.GetLinkWithPassword(downId, pwd, account)
	//} else {
	url, err = driver.GetLink(downId, pwd, account)
	//}
	if err != nil {
		return nil, err
	}
	link := base.Link{
		Url: url,
		Headers: []base.Header{
			{Name: "User-Agent", Value: base.UserAgent},
		},
	}
	return &link, nil
}

func (driver Lanzou) Path(path string, account *model.Account) (*model.File, []model.File, error) {
	path = utils.ParsePath(path)
	log.Debugf("lanzou path: %s", path)
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

//func (driver Lanzou) Proxy(r *http.Request, account *model.Account) {
//	r.Header.Del("Origin")
//}

func (driver Lanzou) Preview(path string, account *model.Account) (interface{}, error) {
	return nil, base.ErrNotSupport
}

func (driver *Lanzou) MakeDir(path string, account *model.Account) error {
	return base.ErrNotImplement
}

func (driver *Lanzou) Move(src string, dst string, account *model.Account) error {
	return base.ErrNotImplement
}

func (driver *Lanzou) Rename(src string, dst string, account *model.Account) error {
	return base.ErrNotImplement
}

func (driver *Lanzou) Copy(src string, dst string, account *model.Account) error {
	return base.ErrNotImplement
}

func (driver *Lanzou) Delete(path string, account *model.Account) error {
	return base.ErrNotImplement
}

func (driver *Lanzou) Upload(file *model.FileStream, account *model.Account) error {
	return base.ErrNotImplement
}

var _ base.Driver = (*Lanzou)(nil)
