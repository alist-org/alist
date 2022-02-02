package webdav

import (
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"path/filepath"
)

type WebDav struct{}

func (driver WebDav) Config() base.DriverConfig {
	return base.DriverConfig{
		Name:          "WebDav",
		OnlyProxy:     true,
		OnlyLocal:     true,
		NoNeedSetLink: true,
		LocalSort:     true,
	}
}

func (driver WebDav) Items() []base.Item {
	return []base.Item{
		{
			Name:     "site_url",
			Label:    "webdav root url",
			Type:     base.TypeString,
			Required: true,
		},
		{
			Name:     "username",
			Label:    "username",
			Type:     base.TypeString,
			Required: true,
		},
		{
			Name:     "password",
			Label:    "password",
			Type:     base.TypeString,
			Required: true,
		},
	}
}

func (driver WebDav) Save(account *model.Account, old *model.Account) error {
	if account == nil {
		return nil
	}
	account.Status = "work"
	_ = model.SaveAccount(account)
	return nil
}

func (driver WebDav) File(path string, account *model.Account) (*model.File, error) {
	if path == "/" {
		return &model.File{
			Id:        "/",
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

func (driver WebDav) Files(path string, account *model.Account) ([]model.File, error) {
	path = utils.ParsePath(path)
	cache, err := base.GetCache(path, account)
	if err == nil {
		files, _ := cache.([]model.File)
		return files, nil
	}
	c := driver.NewClient(account)
	rawFiles, err := c.ReadDir(driver.WebDavPath(path))
	if err != nil {
		return nil, err
	}
	files := make([]model.File, 0)
	if len(rawFiles) == 0 {
		return files, nil
	}
	for _, f := range rawFiles {
		t := f.ModTime()
		file := model.File{
			Name:      f.Name(),
			Size:      f.Size(),
			Driver:    driver.Config().Name,
			UpdatedAt: &t,
		}
		if f.IsDir() {
			file.Type = conf.FOLDER
		} else {
			file.Type = utils.GetFileType(filepath.Ext(f.Name()))
		}
		files = append(files, file)
	}
	_ = base.SetCache(path, files, account)
	return files, nil
}

func (driver WebDav) Link(args base.Args, account *model.Account) (*base.Link, error) {
	path := args.Path
	c := driver.NewClient(account)
	reader, err := c.ReadStream(driver.WebDavPath(path))
	if err != nil {
		return nil, err
	}
	return &base.Link{Data: reader}, nil
}

func (driver WebDav) Path(path string, account *model.Account) (*model.File, []model.File, error) {
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

//func (driver WebDav) Proxy(r *http.Request, account *model.Account) {
//
//}

func (driver WebDav) Preview(path string, account *model.Account) (interface{}, error) {
	return nil, base.ErrNotSupport
}

func (driver WebDav) MakeDir(path string, account *model.Account) error {
	c := driver.NewClient(account)
	err := c.MkdirAll(driver.WebDavPath(path), 0644)
	return err
}

func (driver WebDav) Move(src string, dst string, account *model.Account) error {
	c := driver.NewClient(account)
	err := c.Rename(driver.WebDavPath(src), driver.WebDavPath(dst), true)
	return err
}

func (driver WebDav) Rename(src string, dst string, account *model.Account) error {
	return driver.Move(src, dst, account)
}

func (driver WebDav) Copy(src string, dst string, account *model.Account) error {
	c := driver.NewClient(account)
	err := c.Copy(driver.WebDavPath(src), driver.WebDavPath(dst), true)
	return err
}

func (driver WebDav) Delete(path string, account *model.Account) error {
	c := driver.NewClient(account)
	err := c.RemoveAll(driver.WebDavPath(path))
	return err
}

func (driver WebDav) Upload(file *model.FileStream, account *model.Account) error {
	if file == nil {
		return base.ErrEmptyFile
	}
	c := driver.NewClient(account)
	path := utils.Join(file.ParentPath, file.Name)
	err := c.WriteStream(driver.WebDavPath(path), file, 0644)
	return err
}

var _ base.Driver = (*WebDav)(nil)
