package yandex

import (
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	log "github.com/sirupsen/logrus"
	"net/http"
	"path/filepath"
	"strconv"
)

type Yandex struct{}

func (driver Yandex) Config() base.DriverConfig {
	return base.DriverConfig{
		Name: "Yandex.Disk",
	}
}

func (driver Yandex) Items() []base.Item {
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
		{
			Name:     "order_by",
			Label:    "order_by",
			Type:     base.TypeSelect,
			Default:  "name",
			Values:   "name,path,created,modified,size",
			Required: false,
		},
		{
			Name:     "order_direction",
			Label:    "order_direction",
			Type:     base.TypeSelect,
			Values:   "asc,desc",
			Default:  "asc",
			Required: false,
		},
		{
			Name:     "client_id",
			Label:    "client id",
			Default:  "a78d5a69054042fa936f6c77f9a0ae8b",
			Type:     base.TypeString,
			Required: true,
		},
		{
			Name:     "client_secret",
			Label:    "client secret",
			Default:  "9c119bbb04b346d2a52aa64401936b2b",
			Type:     base.TypeString,
			Required: true,
		},
	}
}

func (driver Yandex) Save(account *model.Account, old *model.Account) error {
	return driver.RefreshToken(account)
}

func (driver Yandex) File(path string, account *model.Account) (*model.File, error) {
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

func (driver Yandex) Files(path string, account *model.Account) ([]model.File, error) {
	path = utils.ParsePath(path)
	cache, err := base.GetCache(path, account)
	if err == nil {
		files, _ := cache.([]model.File)
		return files, nil
	}
	files, err := driver.GetFiles(path, account)
	if err != nil {
		return nil, err
	}
	if len(files) > 0 {
		_ = base.SetCache(path, files, account)
	}
	return files, nil
}

func (driver Yandex) Link(args base.Args, account *model.Account) (*base.Link, error) {
	path := utils.Join(account.RootFolder, args.Path)
	log.Debugln("down path:", path)
	var resp DownResp
	_, err := driver.Request("/download", base.Get, nil, map[string]string{
		"path": path,
	}, nil, nil, &resp, account)
	if err != nil {
		return nil, err
	}
	link := base.Link{
		Url: resp.Href,
	}
	return &link, nil
}

func (driver Yandex) Path(path string, account *model.Account) (*model.File, []model.File, error) {
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

//func (driver Yandex) Proxy(r *http.Request, account *model.Account) {
//
//}

func (driver Yandex) Preview(path string, account *model.Account) (interface{}, error) {
	return nil, base.ErrNotSupport
}

func (driver Yandex) MakeDir(path string, account *model.Account) error {
	path = utils.Join(account.RootFolder, path)
	_, err := driver.Request("", base.Put, nil, map[string]string{
		"path": path,
	}, nil, nil, nil, account)
	return err
}

func (driver Yandex) Move(src string, dst string, account *model.Account) error {
	from := utils.Join(account.RootFolder, src)
	path := utils.Join(account.RootFolder, dst)
	_, err := driver.Request("/move", base.Post, nil, map[string]string{
		"from":      from,
		"path":      path,
		"overwrite": "true",
	}, nil, nil, nil, account)
	return err
}

func (driver Yandex) Rename(src string, dst string, account *model.Account) error {
	return driver.Move(src, dst, account)
}

func (driver Yandex) Copy(src string, dst string, account *model.Account) error {
	from := utils.Join(account.RootFolder, src)
	path := utils.Join(account.RootFolder, dst)
	_, err := driver.Request("/copy", base.Post, nil, map[string]string{
		"from":      from,
		"path":      path,
		"overwrite": "true",
	}, nil, nil, nil, account)
	return err
}

func (driver Yandex) Delete(path string, account *model.Account) error {
	path = utils.Join(account.RootFolder, path)
	_, err := driver.Request("", base.Delete, nil, map[string]string{
		"path": path,
	}, nil, nil, nil, account)
	return err
}

func (driver Yandex) Upload(file *model.FileStream, account *model.Account) error {
	if file == nil {
		return base.ErrEmptyFile
	}
	path := utils.Join(account.RootFolder, file.ParentPath, file.Name)
	var resp UploadResp
	_, err := driver.Request("/upload", base.Get, nil, map[string]string{
		"path":      path,
		"overwrite": "true",
	}, nil, nil, &resp, account)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(resp.Method, resp.Href, file)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Length", strconv.FormatUint(file.Size, 10))
	req.Header.Set("Content-Type", "application/octet-stream")
	res, err := base.HttpClient.Do(req)
	res.Body.Close()
	//res, err := base.RestyClient.R().
	//	SetHeader("Content-Length", strconv.FormatUint(file.Size, 10)).
	//	SetBody(file).Put(resp.Href)
	//log.Debugln(res.Status(), res.String())
	return err
}

var _ base.Driver = (*Yandex)(nil)
