package template

import (
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"path/filepath"
)

type Template struct {
	base.Base
}

func (driver Template) Config() base.DriverConfig {
	return base.DriverConfig{
		Name:          "Template",
		OnlyProxy:     false,
		OnlyLocal:     false,
		ApiProxy:      false,
		NoNeedSetLink: false,
		NoCors:        false,
		LocalSort:     false,
	}
}

func (driver Template) Items() []base.Item {
	// TODO fill need info
	return []base.Item{
		{
			Name:     "refresh_token",
			Label:    "refresh token",
			Type:     base.TypeString,
			Required: true,
		},
		{
			Name:     "root_folder",
			Label:    "root folder path",
			Type:     base.TypeString,
			Default:  "/",
			Required: true,
		},
	}
}

func (driver Template) Save(account *model.Account, old *model.Account) error {
	// TODO test available or init
	return nil
}

func (driver Template) File(path string, account *model.Account) (*model.File, error) {
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

func (driver Template) Files(path string, account *model.Account) ([]model.File, error) {
	path = utils.ParsePath(path)
	cache, err := base.GetCache(path, account)
	if err == nil {
		files, _ := cache.([]model.File)
		return files, nil
	}
	var files []model.File
	// TODO get files
	if err != nil {
		return nil, err
	}
	if len(files) > 0 {
		_ = base.SetCache(path, files, account)
	}
	return files, nil
}

func (driver Template) Link(args base.Args, account *model.Account) (*base.Link, error) {
	// TODO get file link
	return nil, base.ErrNotImplement
}

func (driver Template) Path(path string, account *model.Account) (*model.File, []model.File, error) {
	path = utils.ParsePath(path)
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

// Optional function
//func (driver Template) Preview(path string, account *model.Account) (interface{}, error) {
//	//TODO preview interface if driver support
//	return nil, base.ErrNotImplement
//}
//
//func (driver Template) MakeDir(path string, account *model.Account) error {
//	//TODO make dir
//	return base.ErrNotImplement
//}
//
//func (driver Template) Move(src string, dst string, account *model.Account) error {
//	//TODO move file/dir
//	return base.ErrNotImplement
//}
//
//func (driver Template) Rename(src string, dst string, account *model.Account) error {
//	//TODO rename file/dir
//	return base.ErrNotImplement
//}
//
//func (driver Template) Copy(src string, dst string, account *model.Account) error {
//	//TODO copy file/dir
//	return base.ErrNotImplement
//}
//
//func (driver Template) Delete(path string, account *model.Account) error {
//	//TODO delete file/dir
//	return base.ErrNotImplement
//}
//
//func (driver Template) Upload(file *model.FileStream, account *model.Account) error {
//	//TODO upload file
//	return base.ErrNotImplement
//}

var _ base.Driver = (*Template)(nil)
