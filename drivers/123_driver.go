package drivers

import (
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	url "net/url"
	"path/filepath"
)

type Pan123 struct {}

func (driver Pan123) Config() DriverConfig {
	return DriverConfig{
		Name: "123Pan",
		OnlyProxy: false,
	}
}

func (driver Pan123) Items() []Item {
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
			Required: false,
		},
		{
			Name:     "order_by",
			Label:    "order_by",
			Type:     "select",
			Values:   "name,fileId,updateAt,createAt",
			Required: true,
		},
		{
			Name:     "order_direction",
			Label:    "order_direction",
			Type:     "select",
			Values:   "asc,desc",
			Required: true,
		},
	}
}

func (driver Pan123) Save(account *model.Account, old *model.Account) error {
	if account.RootFolder == "" {
		account.RootFolder = "0"
	}
	err := driver.Login(account)
	return err
}

func (driver Pan123) File(path string, account *model.Account) (*model.File, error) {
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

func (driver Pan123) Files(path string, account *model.Account) ([]model.File, error) {
	path = utils.ParsePath(path)
	var rawFiles []Pan123File
	cache, err := conf.Cache.Get(conf.Ctx, fmt.Sprintf("%s%s", account.Name, path))
	if err == nil {
		rawFiles, _ = cache.([]Pan123File)
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

func (driver Pan123) Link(path string, account *model.Account) (string, error) {
	file, err := driver.GetFile(utils.ParsePath(path), account)
	if err != nil {
		return "", err
	}
	var resp Pan123DownResp
	_, err = pan123Client.R().SetResult(&resp).SetHeader("authorization", "Bearer "+account.AccessToken).
		SetBody(Json{
			"driveId":   0,
			"etag":      file.Etag,
			"fileId":    file.FileId,
			"fileName":  file.FileName,
			"s3keyFlag": file.S3KeyFlag,
			"size":      file.Size,
			"type":      file.Type,
		}).Post("https://www.123pan.com/api/file/download_info")
	if err != nil {
		return "", err
	}
	if resp.Code != 0 {
		if resp.Code == 401 {
			err := driver.Login(account)
			if err != nil {
				return "", err
			}
			return driver.Link(path, account)
		}
		return "", fmt.Errorf(resp.Message)
	}
	u,err := url.Parse(resp.Data.DownloadUrl)
	if err != nil {
		return "", err
	}
	u_ := fmt.Sprintf("https://%s%s",u.Host,u.Path)
	res, err := NoRedirectClient.R().SetQueryParamsFromValues(u.Query()).Get(u_)
	if err != nil {
		return "", err
	}
	log.Debug(res.String())
	if res.StatusCode() == 302 {
		return res.Header().Get("location"), nil
	}
	return resp.Data.DownloadUrl, nil
}

func (driver Pan123) Path(path string, account *model.Account) (*model.File, []model.File, error) {
	path = utils.ParsePath(path)
	log.Debugf("pan123 path: %s", path)
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

func (driver Pan123) Proxy(c *gin.Context, account *model.Account) {
	c.Request.Header.Del("origin")
}

func (driver Pan123) Preview(path string, account *model.Account) (interface{}, error) {
	return nil, NotSupport
}

var _ Driver = (*Pan123)(nil)