package google

import (
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"path/filepath"
)

type GoogleDrive struct{}

func (driver GoogleDrive) Config() base.DriverConfig {
	return base.DriverConfig{
		Name: "GoogleDrive",
		OnlyProxy: true,
	}
}

func (driver GoogleDrive) Items() []base.Item {
	return []base.Item{
		{
			Name:     "client_id",
			Label:    "client id",
			Type:     base.TypeString,
			Required: true,
		},
		{
			Name:     "client_secret",
			Label:    "client secret",
			Type:     base.TypeString,
			Required: true,
		},
		{
			Name:     "refresh_token",
			Label:    "refresh token",
			Type:     base.TypeString,
			Required: true,
		},
		{
			Name:     "root_folder",
			Label:    "root folder file_id",
			Type:     base.TypeString,
			Required: false,
		},
	}
}

func (driver GoogleDrive) Save(account *model.Account, old *model.Account) error {
	account.Proxy = true
	err := driver.RefreshToken(account)
	if err != nil {
		account.Status = err.Error()
		_ = model.SaveAccount(account)
		return err
	}
	if account.RootFolder == "" {
		account.RootFolder = "root"
	}
	account.Status = "work"
	_ = model.SaveAccount(account)
	return nil
}

func (driver GoogleDrive) File(path string, account *model.Account) (*model.File, error) {
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

func (driver GoogleDrive) Files(path string, account *model.Account) ([]model.File, error) {
	path = utils.ParsePath(path)
	var rawFiles []GoogleFile
	cache, err := conf.Cache.Get(conf.Ctx, fmt.Sprintf("%s%s", account.Name, path))
	if err == nil {
		rawFiles, _ = cache.([]GoogleFile)
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
			_ = conf.Cache.Set(conf.Ctx, fmt.Sprintf("%s%s", account.Name, path), rawFiles, nil)
		}
	}
	files := make([]model.File, 0)
	for _, file := range rawFiles {
		files = append(files, *driver.FormatFile(&file))
	}
	return files, nil
}

func (driver GoogleDrive) Link(path string, account *model.Account) (string, error) {
	file, err := driver.File(path, account)
	if err != nil {
		return "", err
	}
	if file.Type == conf.FOLDER {
		return "", base.ErrNotFile
	}
	link := fmt.Sprintf("https://www.googleapis.com/drive/v3/files/%s?includeItemsFromAllDrives=true&supportsAllDrives=true", file.Id)
	var e GoogleError
	_, _ = googleClient.R().SetError(&e).
		SetHeader("Authorization", "Bearer "+account.AccessToken).
		Get(link)
	if e.Error.Code != 0 {
		if e.Error.Code == 401 {
			err = driver.RefreshToken(account)
			if err != nil {
				_ = model.SaveAccount(account)
				return "", err
			}
			return driver.Link(path, account)
		}
		return "", fmt.Errorf("%s: %v", e.Error.Message, e.Error.Errors)
	}
	return link + "&alt=media", nil
}

func (driver GoogleDrive) Path(path string, account *model.Account) (*model.File, []model.File, error) {
	path = utils.ParsePath(path)
	log.Debugf("google path: %s", path)
	file, err := driver.File(path, account)
	if err != nil {
		return nil, nil, err
	}
	if file.Type != conf.FOLDER {
		//file.Url, _ = driver.Link(path, account)
		return file, nil, nil
	}
	files, err := driver.Files(path, account)
	if err != nil {
		return nil, nil, err
	}
	return nil, files, nil
}

func (driver GoogleDrive) Proxy(c *gin.Context, account *model.Account) {
	c.Request.Header.Add("Authorization", "Bearer "+account.AccessToken)
}

func (driver GoogleDrive) Preview(path string, account *model.Account) (interface{}, error) {
	return nil, base.ErrNotSupport
}

func (driver GoogleDrive) MakeDir(path string, account *model.Account) error {
	return base.ErrNotImplement
}

func (driver GoogleDrive) Move(src string, dst string, account *model.Account) error {
	return base.ErrNotImplement
}

func (driver GoogleDrive) Copy(src string, dst string, account *model.Account) error {
	return base.ErrNotImplement
}

func (driver GoogleDrive) Delete(path string, account *model.Account) error {
	return base.ErrNotImplement
}

func (driver GoogleDrive) Upload(file *model.FileStream, account *model.Account) error {
	return base.ErrNotImplement
}

var _ base.Driver = (*GoogleDrive)(nil)