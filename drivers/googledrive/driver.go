package googledrive

import (
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"path/filepath"
)

type GoogleDrive struct{}

var driverName = "GoogleDrive"

func (driver GoogleDrive) Config() drivers.DriverConfig {
	return drivers.DriverConfig{
		OnlyProxy: true,
	}
}

func (driver GoogleDrive) Items() []drivers.Item {
	return []drivers.Item{
		{
			Name:     "client_id",
			Label:    "client id",
			Type:     "string",
			Required: true,
		},
		{
			Name:     "client_secret",
			Label:    "client secret",
			Type:     "string",
			Required: true,
		},
		{
			Name:     "refresh_token",
			Label:    "refresh token",
			Type:     "string",
			Required: true,
		},
		{
			Name:     "root_folder",
			Label:    "root folder file_id",
			Type:     "string",
			Required: true,
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
			Driver:    driverName,
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
	return nil, drivers.PathNotFound
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
		return "", drivers.NotFile
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
	return nil, nil
}
