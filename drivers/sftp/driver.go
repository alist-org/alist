package template

import (
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"io"
	"path"
	"path/filepath"
)

type SFTP struct {
}

func (driver SFTP) Config() base.DriverConfig {
	return base.DriverConfig{
		Name:      "SFTP",
		OnlyProxy: true,
		OnlyLocal: true,
		LocalSort: true,
	}
}

func (driver SFTP) Items() []base.Item {
	// TODO fill need info
	return []base.Item{
		{
			Name:     "site_url",
			Label:    "ip/host",
			Type:     base.TypeString,
			Required: true,
		},
		{
			Name:     "limit",
			Label:    "port",
			Type:     base.TypeNumber,
			Required: true,
			Default:  "22",
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
			Default:  "/",
			Required: true,
		},
	}
}

func (driver SFTP) Save(account *model.Account, old *model.Account) error {
	if old != nil {
		clientsMap.Lock()
		defer clientsMap.Unlock()
		delete(clientsMap.clients, old.Name)
	}
	if account == nil {
		return nil
	}
	_, err := GetClient(account)
	if err != nil {
		account.Status = err.Error()
	} else {
		account.Status = "work"
	}
	_ = model.SaveAccount(account)
	return err
}

func (driver SFTP) File(path string, account *model.Account) (*model.File, error) {
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

func (driver SFTP) Files(path string, account *model.Account) ([]model.File, error) {
	path = utils.ParsePath(path)
	remotePath := utils.Join(account.RootFolder, path)
	cache, err := base.GetCache(path, account)
	if err == nil {
		files, _ := cache.([]model.File)
		return files, nil
	}
	client, err := GetClient(account)
	if err != nil {
		return nil, err
	}
	files := make([]model.File, 0)
	rawFiles, err := client.Files(remotePath)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(rawFiles); i++ {
		files = append(files, driver.formatFile(rawFiles[i]))
	}
	if len(files) > 0 {
		_ = base.SetCache(path, files, account)
	}
	return files, nil
}

func (driver SFTP) Link(args base.Args, account *model.Account) (*base.Link, error) {
	client, err := GetClient(account)
	if err != nil {
		return nil, err
	}
	remoteFileName := utils.Join(account.RootFolder, args.Path)
	remoteFile, err := client.Open(remoteFileName)
	if err != nil {
		return nil, err
	}
	return &base.Link{
		Data: remoteFile,
	}, nil
}

func (driver SFTP) Path(path string, account *model.Account) (*model.File, []model.File, error) {
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

func (driver SFTP) Preview(path string, account *model.Account) (interface{}, error) {
	//TODO preview interface if driver support
	return nil, base.ErrNotImplement
}

func (driver SFTP) MakeDir(path string, account *model.Account) error {
	client, err := GetClient(account)
	if err != nil {
		return err
	}
	return client.MkdirAll(utils.Join(account.RootFolder, path))
}

func (driver SFTP) Move(src string, dst string, account *model.Account) error {
	return driver.Rename(src, dst, account)
}

func (driver SFTP) Rename(src string, dst string, account *model.Account) error {
	client, err := GetClient(account)
	if err != nil {
		return err
	}
	return client.Rename(utils.Join(account.RootFolder, src), utils.Join(account.RootFolder, dst))
}

func (driver SFTP) Copy(src string, dst string, account *model.Account) error {
	return base.ErrNotSupport
}

func (driver SFTP) Delete(path string, account *model.Account) error {
	client, err := GetClient(account)
	if err != nil {
		return err
	}
	return client.remove(utils.Join(account.RootFolder, path))
}

func (driver SFTP) Upload(file *model.FileStream, account *model.Account) error {
	if file == nil {
		return base.ErrEmptyFile
	}
	client, err := GetClient(account)
	if err != nil {
		return err
	}
	dstFile, err := client.Create(path.Join(account.RootFolder, file.ParentPath, file.Name))
	if err != nil {
		return err
	}
	defer func() {
		_ = dstFile.Close()
	}()
	_, err = io.Copy(dstFile, file)
	return err
}

var _ base.Driver = (*SFTP)(nil)
