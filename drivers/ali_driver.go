package drivers

import (
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"path/filepath"
)

type AliDrive struct{}

func (driver AliDrive) Config() DriverConfig {
	return DriverConfig{
		Name: "AliDrive",
		OnlyProxy: false,
	}
}

func (driver AliDrive) Items() []Item {
	return []Item{
		{
			Name: "order_by",
			Label: "order_by",
			Type: "select",
			Values: "name,size,updated_at,created_at",
			Required: false,
		},
		{
			Name: "order_direction",
			Label: "order_direction",
			Type: "select",
			Values: "ASC,DESC",
			Required: false,
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
			Required: false,
		},
		{
			Name:        "limit",
			Label:       "limit",
			Type:        "number",
			Required:    false,
			Description: ">0 and <=200",
		},
	}
}

func (driver AliDrive) Save(account *model.Account, old *model.Account) error {
	if old != nil {
		conf.Cron.Remove(cron.EntryID(old.CronId))
	}
	if account.RootFolder == "" {
		account.RootFolder = "root"
	}
	if account.Limit == 0 {
		account.Limit = 200
	}
	err := driver.RefreshToken(account)
	if err != nil {
		return err
	}
	var resp Json
	_, _ = aliClient.R().SetResult(&resp).
		SetBody("{}").
		SetHeader("authorization", "Bearer\t"+account.AccessToken).
		Post("https://api.aliyundrive.com/v2/user/get")
	log.Debugf("user info: %+v", resp)
	account.DriveId = resp["default_drive_id"].(string)
	cronId, err := conf.Cron.AddFunc("@every 2h", func() {
		name := account.Name
		log.Debugf("ali account name: %s", name)
		newAccount, ok := model.GetAccount(name)
		log.Debugf("ali account: %+v", newAccount)
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

func (driver AliDrive) File(path string, account *model.Account) (*model.File, error) {
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
	return nil, PathNotFound
}

func (driver AliDrive) Files(path string, account *model.Account) ([]model.File, error) {
	path = utils.ParsePath(path)
	var rawFiles []AliFile
	cache, err := conf.Cache.Get(conf.Ctx, fmt.Sprintf("%s%s", account.Name, path))
	if err == nil {
		rawFiles, _ = cache.([]AliFile)
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

func (driver AliDrive) Link(path string, account *model.Account) (string, error) {
	file, err := driver.File(path, account)
	if err != nil {
		return "", err
	}
	var resp Json
	var e AliRespError
	_, err = aliClient.R().SetResult(&resp).
		SetError(&e).
		SetHeader("authorization", "Bearer\t"+account.AccessToken).
		SetBody(Json{
			"drive_id":   account.DriveId,
			"file_id":    file.Id,
			"expire_sec": 14400,
		}).Post("https://api.aliyundrive.com/v2/file/get_download_url")
	if err != nil {
		return "", err
	}
	if e.Code != "" {
		if e.Code == "AccessTokenInvalid" {
			err = driver.RefreshToken(account)
			if err != nil {
				return "", err
			} else {
				_ = model.SaveAccount(account)
				return driver.Link(path, account)
			}
		}
		return "", fmt.Errorf("%s", e.Message)
	}
	return resp["url"].(string), nil
}

func (driver AliDrive) Path(path string, account *model.Account) (*model.File, []model.File, error) {
	path = utils.ParsePath(path)
	log.Debugf("ali path: %s", path)
	file, err := driver.File(path, account)
	if err != nil {
		return nil, nil, err
	}
	if file.Type != conf.FOLDER {
		file.Url, _ = driver.Link(path, account)
		return file, nil, nil
	}
	files, err := driver.Files(path, account)
	if err != nil {
		return nil, nil, err
	}
	return nil, files, nil
}

func (driver AliDrive) Proxy(c *gin.Context, account *model.Account) {
	c.Request.Header.Del("Origin")
	c.Request.Header.Set("Referer", "https://www.aliyundrive.com/")
}

func (driver AliDrive) Preview(path string, account *model.Account) (interface{}, error) {
	file, err := driver.GetFile(path, account)
	if err != nil {
		return nil, err
	}
	// office
	var resp Json
	var e AliRespError
	var url string
	req := Json{
		"drive_id": account.DriveId,
		"file_id":  file.FileId,
	}
	switch file.Category {
	case "doc":
		{
			url = "https://api.aliyundrive.com/v2/file/get_office_preview_url"
			req["access_token"] = account.AccessToken
		}
	case "video":
		{
			url = "https://api.aliyundrive.com/v2/file/get_video_preview_play_info"
			req["category"] = "live_transcoding"
		}
	default:
		return nil, fmt.Errorf("don't support")
	}
	_, err = aliClient.R().SetResult(&resp).SetError(&e).
		SetHeader("authorization", "Bearer\t"+account.AccessToken).
		SetBody(req).Post(url)
	if err != nil {
		return nil, err
	}
	if e.Code != "" {
		return nil, fmt.Errorf("%s", e.Message)
	}
	return resp, nil
}

var _ Driver = (*AliDrive)(nil)