package drivers

import (
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"path/filepath"
)

type Cloud189 struct {}

func (driver Cloud189) Config() DriverConfig {
	return DriverConfig{
		Name: "189Cloud",
		OnlyProxy: false,
	}
}

func (driver Cloud189) Items() []Item {
	return []Item{
		{
			Name:        "username",
			Label:       "username",
			Type:        "string",
			Required:    true,
			Description: "account username/phone number",
		},
		{
			Name:        "password",
			Label:       "password",
			Type:        "string",
			Required:    true,
			Description: "account password",
		},
		{
			Name:     "root_folder",
			Label:    "root folder file_id",
			Type:     "string",
			Required: true,
		},
		{
			Name:     "order_by",
			Label:    "order_by",
			Type:     "select",
			Values:   "name,size,lastOpTime,createdDate",
			Required: true,
		},
		{
			Name:     "order_direction",
			Label:    "desc",
			Type:     "select",
			Values:   "true,false",
			Required: true,
		},
	}
}

func (driver Cloud189) Save(account *model.Account, old *model.Account) error {
	if old != nil && old.Name != account.Name {
		delete(client189Map, old.Name)
	}
	if err := driver.Login(account); err != nil {
		account.Status = err.Error()
		_ = model.SaveAccount(account)
		return err
	}
	account.Status = "work"
	err := model.SaveAccount(account)
	if err != nil {
		return err
	}
	return nil
}

func (driver Cloud189) File(path string, account *model.Account) (*model.File, error) {
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

func (driver Cloud189) Files(path string, account *model.Account) ([]model.File, error) {
	path = utils.ParsePath(path)
	var rawFiles []Cloud189File
	cache, err := conf.Cache.Get(conf.Ctx, fmt.Sprintf("%s%s", account.Name, path))
	if err == nil {
		rawFiles, _ = cache.([]Cloud189File)
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

func (driver Cloud189) Link(path string, account *model.Account) (string, error) {
	file, err := driver.File(utils.ParsePath(path), account)
	if err != nil {
		return "", err
	}
	if file.Type == conf.FOLDER {
		return "", NotFile
	}
	client, ok := client189Map[account.Name]
	if !ok {
		return "", fmt.Errorf("can't find [%s] client", account.Name)
	}
	var e Cloud189Error
	var resp Cloud189Down
	_, err = client.R().SetResult(&resp).SetError(&e).
		SetHeader("Accept", "application/json;charset=UTF-8").
		SetQueryParams(map[string]string{
			"noCache": random(),
			"fileId":  file.Id,
		}).Get("https://cloud.189.cn/api/open/file/getFileDownloadUrl.action")
	if err != nil {
		return "", err
	}
	if e.ErrorCode != "" {
		if e.ErrorCode == "InvalidSessionKey" {
			err = driver.Login(account)
			if err != nil {
				return "", err
			}
			return driver.Link(path, account)
		}
	}
	if resp.ResCode != 0 {
		return "", fmt.Errorf(resp.ResMessage)
	}
	res, err := NoRedirectClient.R().Get(resp.FileDownloadUrl)
	if err != nil {
		return "", err
	}
	if res.StatusCode() == 302 {
		return res.Header().Get("location"), nil
	}
	return resp.FileDownloadUrl, nil
}

func (driver Cloud189) Path(path string, account *model.Account) (*model.File, []model.File, error) {
	path = utils.ParsePath(path)
	log.Debugf("189 path: %s", path)
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

func (driver Cloud189) Proxy(ctx *gin.Context, account *model.Account) {
	ctx.Request.Header.Del("Origin")
}

func (driver Cloud189) Preview(path string, account *model.Account) (interface{}, error) {
	return nil, NotSupport
}

var _ Driver = (*Cloud189)(nil)