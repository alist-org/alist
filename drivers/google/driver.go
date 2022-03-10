package google

import (
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"path/filepath"
)

type GoogleDrive struct{}

func (driver GoogleDrive) Config() base.DriverConfig {
	return base.DriverConfig{
		Name:          "GoogleDrive",
		OnlyProxy:     true,
		ApiProxy:      true,
		NoNeedSetLink: true,
	}
}

func (driver GoogleDrive) Items() []base.Item {
	return []base.Item{
		{
			Name:     "client_id",
			Label:    "client id",
			Type:     base.TypeString,
			Required: true,
			Default:  "202264815644.apps.googleusercontent.com",
		},
		{
			Name:     "client_secret",
			Label:    "client secret",
			Type:     base.TypeString,
			Required: true,
			Default:  "X4Z3ca8xfWDb1Voo-F9a7ZxJ",
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
		{
			Name:        "order_by",
			Label:       "order_by",
			Type:        base.TypeString,
			Required:    false,
			Description: "such as: folder,name,modifiedTime",
		},
		{
			Name:     "order_direction",
			Label:    "order_direction",
			Type:     base.TypeSelect,
			Values:   "asc,desc",
			Required: false,
		},
	}
}

func (driver GoogleDrive) Save(account *model.Account, old *model.Account) error {
	if account == nil {
		return nil
	}
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
	var rawFiles []File
	cache, err := base.GetCache(path, account)
	if err == nil {
		rawFiles, _ = cache.([]File)
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
			_ = base.SetCache(path, rawFiles, account)
		}
	}
	files := make([]model.File, 0)
	for _, file := range rawFiles {
		files = append(files, *driver.FormatFile(&file, account))
	}
	return files, nil
}

func (driver GoogleDrive) Link(args base.Args, account *model.Account) (*base.Link, error) {
	file, err := driver.File(args.Path, account)
	if err != nil {
		return nil, err
	}
	if file.Type == conf.FOLDER {
		return nil, base.ErrNotFile
	}
	url := fmt.Sprintf("https://www.googleapis.com/drive/v3/files/%s?includeItemsFromAllDrives=true&supportsAllDrives=true", file.Id)
	_, err = driver.Request(url, base.Get, nil, nil, nil, nil, nil, account)
	if err != nil {
		return nil, err
	}
	link := base.Link{
		Url: url + "&alt=media",
		Headers: []base.Header{
			{
				Name:  "Authorization",
				Value: "Bearer " + account.AccessToken,
			},
		},
	}
	return &link, nil
}

func (driver GoogleDrive) Path(path string, account *model.Account) (*model.File, []model.File, error) {
	path = utils.ParsePath(path)
	log.Debugf("google path: %s", path)
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

//func (driver GoogleDrive) Proxy(r *http.Request, account *model.Account) {
//	r.Header.Add("Authorization", "Bearer "+account.AccessToken)
//}

func (driver GoogleDrive) Preview(path string, account *model.Account) (interface{}, error) {
	return nil, base.ErrNotSupport
}

func (driver GoogleDrive) MakeDir(path string, account *model.Account) error {
	parentFile, err := driver.File(utils.Dir(path), account)
	if err != nil {
		return err
	}
	data := base.Json{
		"name":     utils.Base(path),
		"parents":  []string{parentFile.Id},
		"mimeType": "application/vnd.google-apps.folder",
	}
	_, err = driver.Request("https://www.googleapis.com/drive/v3/files", base.Post, nil, nil, nil, &data, nil, account)
	return err
}

func (driver GoogleDrive) Move(src string, dst string, account *model.Account) error {
	srcFile, err := driver.File(src, account)
	url := "https://www.googleapis.com/drive/v3/files/" + srcFile.Id
	if err != nil {
		return err
	}
	dstParentFile, err := driver.File(utils.Dir(dst), account)
	if err != nil {
		return err
	}
	query := map[string]string{
		"addParents":    dstParentFile.Id,
		"removeParents": "root",
	}
	_, err = driver.Request(url, base.Patch, nil, query, nil, nil, nil, account)
	return err
}

func (driver GoogleDrive) Rename(src string, dst string, account *model.Account) error {
	srcFile, err := driver.File(src, account)
	url := "https://www.googleapis.com/drive/v3/files/" + srcFile.Id
	if err != nil {
		return err
	}
	data := base.Json{
		"name": utils.Base(dst),
	}
	_, err = driver.Request(url, base.Patch, nil, nil, nil, &data, nil, account)
	return err
}

func (driver GoogleDrive) Copy(src string, dst string, account *model.Account) error {
	return base.ErrNotSupport
}

func (driver GoogleDrive) Delete(path string, account *model.Account) error {
	file, err := driver.File(path, account)
	url := "https://www.googleapis.com/drive/v3/files/" + file.Id
	if err != nil {
		return err
	}
	_, err = driver.Request(url, base.Delete, nil, nil, nil, nil, nil, account)
	return err
}

func (driver GoogleDrive) Upload(file *model.FileStream, account *model.Account) error {
	if file == nil {
		return base.ErrEmptyFile
	}
	parentFile, err := driver.File(file.ParentPath, account)
	if err != nil {
		return err
	}
	data := base.Json{
		"name":    file.Name,
		"parents": []string{parentFile.Id},
	}
	var e Error
	url := "https://www.googleapis.com/upload/drive/v3/files?uploadType=resumable&supportsAllDrives=true"
	if account.APIProxyUrl != "" {
		url = fmt.Sprintf("%s/%s", account.APIProxyUrl, url)
	}
	res, err := base.NoRedirectClient.R().SetHeader("Authorization", "Bearer "+account.AccessToken).
		SetError(&e).SetBody(data).
		Post(url)
	if err != nil {
		return err
	}
	if e.Error.Code != 0 {
		if e.Error.Code == 401 {
			err = driver.RefreshToken(account)
			if err != nil {
				_ = model.SaveAccount(account)
				return err
			}
			return driver.Upload(file, account)
		}
		return fmt.Errorf("%s: %v", e.Error.Message, e.Error.Errors)
	}
	putUrl := res.Header().Get("location")
	byteData, _ := ioutil.ReadAll(file)
	_, err = driver.Request(putUrl, base.Put, nil, nil, nil, byteData, nil, account)
	return err
}

var _ base.Driver = (*GoogleDrive)(nil)
