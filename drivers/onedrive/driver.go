package onedrive

import (
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"path/filepath"
)

type Onedrive struct{}

var driverName = "Onedrive"

func (driver Onedrive) Items() []drivers.Item {
	return []drivers.Item{
		{
			Name:        "proxy",
			Label:       "proxy",
			Type:        "bool",
			Required:    true,
			Description: "allow proxy",
		},
		{
			Name:        "zone",
			Label:       "zone",
			Type:        "select",
			Required:    true,
			Values:      "global,cn,us,de",
			Description: "",
		},
		{
			Name:     "onedrive_type",
			Label:    "onedrive type",
			Type:     "select",
			Required: true,
			Values:   "onedrive,sharepoint",
		},
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
			Name:     "redirect_uri",
			Label:    "redirect uri",
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
			Name:     "site_id",
			Label:    "site id",
			Type:     "string",
			Required: false,
		},
		{
			Name:     "root_folder",
			Label:    "root folder path",
			Type:     "string",
			Required: false,
		},
		{
			Name:     "order_by",
			Label:    "order_by",
			Type:     "select",
			Values:   "name,size,lastModifiedDateTime",
			Required: false,
		},
		{
			Name:     "order_direction",
			Label:    "order_direction",
			Type:     "select",
			Values:   "asc,desc",
			Required: false,
		},
	}
}

func (driver Onedrive) Save(account *model.Account, old *model.Account) error {
	_, ok := onedriveHostMap[account.Zone]
	if !ok {
		return fmt.Errorf("no [%s] zone", account.Zone)
	}
	if old != nil {
		conf.Cron.Remove(cron.EntryID(old.CronId))
	}
	account.RootFolder = utils.ParsePath(account.RootFolder)
	err := driver.RefreshToken(account)
	if err != nil {
		return err
	}
	cronId, err := conf.Cron.AddFunc("@every 1h", func() {
		name := account.Name
		log.Debugf("onedrive account name: %s", name)
		newAccount, ok := model.GetAccount(name)
		log.Debugf("onedrive account: %+v", newAccount)
		if !ok {
			return
		}
		err = driver.RefreshToken(&newAccount)
		_ = model.SaveAccount(&newAccount)
	})
	if err != nil {
		return err
	}
	account.CronId = int(cronId)
	err = model.SaveAccount(account)
	if err != nil {
		return err
	}
	return nil
}

func (driver Onedrive) File(path string, account *model.Account) (*model.File, error) {
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

func (driver Onedrive) Files(path string, account *model.Account) ([]model.File, error) {
	path = utils.ParsePath(path)
	cache, err := conf.Cache.Get(conf.Ctx, fmt.Sprintf("%s%s", account.Name, path))
	if err == nil {
		files, _ := cache.([]model.File)
		return files, nil
	}
	rawFiles, err := driver.GetFiles(account, path)
	if err != nil {
		return nil, err
	}
	files := make([]model.File, 0)
	for _, file := range rawFiles {
		files = append(files, *driver.FormatFile(&file))
	}
	_ = conf.Cache.Set(conf.Ctx, fmt.Sprintf("%s%s", account.Name, path), files, nil)
	return files, nil
}

func (driver Onedrive) Link(path string, account *model.Account) (string, error) {
	file, err := driver.GetFile(account, path)
	if err != nil {
		return "", err
	}
	if file.File.MimeType == "" {
		return "", fmt.Errorf("can't down folder")
	}
	return file.Url, nil
}

func (driver Onedrive) Path(path string, account *model.Account) (*model.File, []model.File, error) {
	log.Debugf("onedrive path: %s", path)
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

func (driver Onedrive) Proxy(c *gin.Context, account *model.Account) {
	c.Request.Header.Del("Origin")
}

func (driver Onedrive) Preview(path string, account *model.Account) (interface{}, error) {
	return nil, nil
}
