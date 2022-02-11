package teambition

import (
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	log "github.com/sirupsen/logrus"
	"path/filepath"
)

type Teambition struct{}

func (driver Teambition) Config() base.DriverConfig {
	return base.DriverConfig{
		Name: "Teambition",
	}
}

func (driver Teambition) Items() []base.Item {
	return []base.Item{
		{
			Name:     "internal_type",
			Label:    "Teambition type",
			Type:     base.TypeSelect,
			Required: true,
			Values:   "China,International",
		},
		{
			Name:        "access_token",
			Label:       "Cookie",
			Type:        base.TypeString,
			Required:    true,
			Description: "Unknown expiration time",
		},
		{
			Name:     "zone",
			Label:    "Project id",
			Type:     base.TypeString,
			Required: true,
		},
		{
			Name:     "root_folder",
			Label:    "root folder file_id",
			Type:     base.TypeString,
			Required: true,
		},
		{
			Name:     "order_by",
			Label:    "order_by",
			Type:     base.TypeSelect,
			Values:   "fileName,fileSize,updated,created",
			Required: true,
			Default:  "fileName",
		},
		{
			Name:     "order_direction",
			Label:    "order_direction",
			Type:     base.TypeSelect,
			Values:   "Asc,Desc",
			Required: true,
			Default:  "Asc",
		},
	}
}

func (driver Teambition) Save(account *model.Account, old *model.Account) error {
	if account == nil {
		return nil
	}
	_, err := driver.Request("/api/v2/roles", base.Get, nil, nil, nil, nil, nil, account)
	return err
}

func (driver Teambition) File(path string, account *model.Account) (*model.File, error) {
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

func (driver Teambition) Files(path string, account *model.Account) ([]model.File, error) {
	path = utils.ParsePath(path)
	var files []model.File
	cache, err := base.GetCache(path, account)
	if err == nil {
		files, _ = cache.([]model.File)
	} else {
		file, err := driver.File(path, account)
		if err != nil {
			return nil, err
		}
		files, err = driver.GetFiles(file.Id, account)
		if err != nil {
			return nil, err
		}
		if len(files) > 0 {
			_ = base.SetCache(path, files, account)
		}
	}
	return files, nil
}

func (driver Teambition) Link(args base.Args, account *model.Account) (*base.Link, error) {
	path := args.Path
	file, err := driver.File(path, account)
	if err != nil {
		return nil, err
	}
	url := file.Url
	res, err := base.NoRedirectClient.R().Get(url)
	if res.StatusCode() == 302 {
		url = res.Header().Get("location")
	}
	return &base.Link{Url: url}, nil
}

func (driver Teambition) Path(path string, account *model.Account) (*model.File, []model.File, error) {
	path = utils.ParsePath(path)
	log.Debugf("teambition path: %s", path)
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

//func (driver Teambition) Proxy(r *http.Request, account *model.Account) {
//
//}

func (driver Teambition) Preview(path string, account *model.Account) (interface{}, error) {
	return nil, base.ErrNotSupport
}

func (driver Teambition) MakeDir(path string, account *model.Account) error {
	parentFile, err := driver.File(utils.Dir(path), account)
	if err != nil {
		return err
	}
	data := base.Json{
		"objectType":     "collection",
		"_projectId":     account.Zone,
		"_creatorId":     "",
		"created":        "",
		"updated":        "",
		"title":          utils.Base(path),
		"color":          "blue",
		"description":    "",
		"workCount":      0,
		"collectionType": "",
		"recentWorks":    []interface{}{},
		"_parentId":      parentFile.Id,
		"subCount":       nil,
	}
	_, err = driver.Request("/api/collections", base.Post, nil, nil, nil, &data, nil, account)
	return err
}

func (driver Teambition) Move(src string, dst string, account *model.Account) error {
	srcFile, err := driver.File(src, account)
	if err != nil {
		return err
	}
	dstParentFile, err := driver.File(utils.Dir(dst), account)
	if err != nil {
		return err
	}
	pre := "/api/works/"
	if srcFile.IsDir() {
		pre = "/api/collections/"
	}
	_, err = driver.Request(pre+srcFile.Id+"/move", base.Put, nil, nil, nil, &base.Json{
		"_parentId": dstParentFile.Id,
	}, nil, account)
	return err
}

func (driver Teambition) Rename(src string, dst string, account *model.Account) error {
	srcFile, err := driver.File(src, account)
	if err != nil {
		return err
	}
	pre := "/api/works/"
	data := base.Json{
		"fileName": utils.Base(dst),
	}
	if srcFile.IsDir() {
		pre = "/api/collections/"
		data = base.Json{
			"title": utils.Base(dst),
		}
	}
	_, err = driver.Request(pre+srcFile.Id, base.Put, nil, nil, nil, &data, nil, account)
	return err
}

func (driver Teambition) Copy(src string, dst string, account *model.Account) error {
	srcFile, err := driver.File(src, account)
	if err != nil {
		return err
	}
	dstParentFile, err := driver.File(utils.Dir(dst), account)
	if err != nil {
		return err
	}
	pre := "/api/works/"
	if srcFile.IsDir() {
		pre = "/api/collections/"
	}
	_, err = driver.Request(pre+srcFile.Id+"/fork", base.Put, nil, nil, nil, &base.Json{
		"_parentId": dstParentFile.Id,
	}, nil, account)
	return err
}

func (driver Teambition) Delete(path string, account *model.Account) error {
	srcFile, err := driver.File(path, account)
	if err != nil {
		return err
	}
	pre := "/api/works/"
	if srcFile.IsDir() {
		pre = "/api/collections/"
	}
	_, err = driver.Request(pre+srcFile.Id+"/archive", base.Post, nil, nil, nil, nil, nil, account)
	return err
}

func (driver Teambition) Upload(file *model.FileStream, account *model.Account) error {
	if file == nil {
		return base.ErrEmptyFile
	}
	parentFile, err := driver.File(file.ParentPath, account)
	if !parentFile.IsDir() {
		return base.ErrNotFolder
	}
	if err != nil {
		return err
	}
	res, err := driver.Request("/projects", base.Get, nil, nil, nil, nil, nil, account)
	if err != nil {
		return err
	}
	token := GetBetweenStr(string(res), "strikerAuth&quot;:&quot;", "&quot;,&quot;phoneForLogin")
	var newFile *FileUpload
	if file.Size <= 20971520 {
		// post upload
		newFile, err = driver.upload(file, token, account)
	} else {
		// chunk upload
		//err = base.ErrNotImplement
		newFile, err = driver.chunkUpload(file, token, account)
	}
	if err != nil {
		return err
	}
	return driver.finishUpload(newFile, parentFile.Id, account)
}

var _ base.Driver = (*Teambition)(nil)
