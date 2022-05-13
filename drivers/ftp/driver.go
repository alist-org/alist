package ftp

import (
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/jlaffaye/ftp"
	log "github.com/sirupsen/logrus"
	"path/filepath"
)

type FTP struct{}

func (driver FTP) Config() base.DriverConfig {
	return base.DriverConfig{
		Name:          "FTP",
		OnlyProxy:     true,
		OnlyLocal:     true,
		NoNeedSetLink: true,
		LocalSort:     true,
	}
}

func (driver FTP) Items() []base.Item {
	return []base.Item{
		{
			Name:     "site_url",
			Label:    "ftp host url",
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
		{
			Name:     "root_folder",
			Label:    "root folder path",
			Type:     base.TypeString,
			Required: false,
		},
	}
}

func (driver FTP) Save(account *model.Account, old *model.Account) error {
	if old != nil {
		conn, ok := connMap[old.Name]
		if ok {
			err := conn.Quit()
			log.Error("ftp:", err)
			delete(connMap, old.Name)
		}
	}
	if account == nil {
		return nil
	}
	if account.RootFolder == "" {
		account.RootFolder = "/"
	}
	_, err := driver.Login(account)
	if err != nil {
		account.Status = err.Error()
	} else {
		account.Status = "work"
	}
	_ = model.SaveAccount(account)
	return err
}

func (driver FTP) File(path string, account *model.Account) (*model.File, error) {
	log.Debugf("file: %s", path)
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

func (driver FTP) Files(path string, account *model.Account) ([]model.File, error) {
	log.Debugf("files: %s", path)
	path = utils.ParsePath(path)
	cache, err := base.GetCache(path, account)
	if err == nil {
		files, _ := cache.([]model.File)
		return files, nil
	}
	realPath := utils.Join(account.RootFolder, path)
	conn, err := driver.Login(account)
	if err != nil {
		return nil, err
	}
	//defer func() { _ = conn.Quit() }()
	entries, err := conn.List(realPath)
	if err != nil {
		return nil, err
	}
	res := make([]model.File, 0)
	for i, _ := range entries {
		entry := entries[i]
		if entry.Name == "." || entry.Name == ".." {
			continue
		}
		f := model.File{
			Name:      entry.Name,
			Size:      int64(entry.Size),
			UpdatedAt: &entry.Time,
			Driver:    driver.Config().Name,
		}
		if entry.Type == ftp.EntryTypeFolder {
			f.Type = conf.FOLDER
		} else {
			f.Type = utils.GetFileType(filepath.Ext(entry.Name))
		}
		res = append(res, f)
	}
	if len(res) > 0 {
		_ = base.SetCache(path, res, account)
	}
	return res, nil
}

func (driver FTP) Link(args base.Args, account *model.Account) (*base.Link, error) {
	path := args.Path
	path = utils.ParsePath(path)
	realPath := utils.Join(account.RootFolder, path)
	conn, err := driver.Login(account)
	if err != nil {
		return nil, err
	}
	//defer func() { _ = conn.Quit() }()
	resp, err := conn.Retr(realPath)
	if err != nil {
		return nil, err
	}
	//defer func() { _ = resp.Close() }()
	//data, err := ioutil.ReadAll(resp)
	//if err != nil {
	//	return nil, err
	//}
	return &base.Link{
		Data: resp,
	}, nil
}

func (driver FTP) Path(path string, account *model.Account) (*model.File, []model.File, error) {
	log.Debugf("ftp path: %s", path)
	file, err := driver.File(path, account)
	if err != nil {
		return nil, nil, err
	}
	if !file.IsDir() {
		//file.Url, _ = driver.Link(path, account)
		return file, nil, nil
	}
	files, err := driver.Files(path, account)
	if err != nil {
		return nil, nil, err
	}
	return nil, files, nil
}

//func (driver FTP) Proxy(r *http.Request, account *model.Account) {
//
//}

func (driver FTP) Preview(path string, account *model.Account) (interface{}, error) {
	return nil, base.ErrNotSupport
}

func (driver FTP) MakeDir(path string, account *model.Account) error {
	path = utils.ParsePath(path)
	realPath := utils.Join(account.RootFolder, path)
	conn, err := driver.Login(account)
	if err != nil {
		return err
	}
	//defer func() { _ = conn.Quit() }()
	err = conn.MakeDir(realPath)
	return err
}

func (driver FTP) Move(src string, dst string, account *model.Account) error {
	realSrc := utils.Join(account.RootFolder, src)
	realDst := utils.Join(account.RootFolder, dst)
	conn, err := driver.Login(account)
	if err != nil {
		return err
	}
	//defer func() { _ = conn.Quit() }()
	err = conn.Rename(realSrc, realDst)
	return err
}

func (driver FTP) Rename(src string, dst string, account *model.Account) error {
	return driver.Move(src, dst, account)
}

func (driver FTP) Copy(src string, dst string, account *model.Account) error {
	return base.ErrNotSupport
}

func (driver FTP) Delete(path string, account *model.Account) error {
	path = utils.ParsePath(path)
	file, err := driver.File(path, account)
	if err != nil {
		return err
	}
	realPath := utils.Join(account.RootFolder, path)
	conn, err := driver.Login(account)
	if err != nil {
		return err
	}
	//defer func() { _ = conn.Quit() }()
	if file.IsDir() {
		err = conn.RemoveDirRecur(realPath)
	} else {
		err = conn.Delete(realPath)
	}
	return err
}

func (driver FTP) Upload(file *model.FileStream, account *model.Account) error {
	if file == nil {
		return base.ErrEmptyFile
	}
	realPath := utils.Join(account.RootFolder, file.ParentPath, file.Name)
	conn, err := driver.Login(account)
	if err != nil {
		return err
	}
	//defer func() { _ = conn.Quit() }()
	err = conn.Stor(realPath, file)
	return err
}

var _ base.Driver = (*FTP)(nil)
