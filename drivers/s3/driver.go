package s3

import (
	"bytes"
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"time"
)

type S3 struct {
}

func (driver S3) Config() base.DriverConfig {
	return base.DriverConfig{
		Name:      "S3",
		LocalSort: true,
	}
}

func (driver S3) Items() []base.Item {
	return []base.Item{
		{
			Name:     "bucket",
			Label:    "Bucket",
			Type:     base.TypeString,
			Required: true,
		},
		{
			Name:     "endpoint",
			Label:    "Endpoint",
			Type:     base.TypeString,
			Required: true,
		},
		{
			Name:     "region",
			Label:    "Region",
			Type:     base.TypeString,
			Required: true,
		},
		{
			Name:     "access_key",
			Label:    "Access Key",
			Type:     base.TypeString,
			Required: true,
		},
		{
			Name:     "access_secret",
			Label:    "Access Secret",
			Type:     base.TypeString,
			Required: true,
		},
		{
			Name:     "root_folder",
			Label:    "root folder path",
			Type:     base.TypeString,
			Required: false,
		},
		{
			Name:  "custom_host",
			Label: "Custom Host",
			Type:  base.TypeString,
		},
		{
			Name:        "limit",
			Label:       "Sign url expire time(hours)",
			Type:        base.TypeNumber,
			Description: "default 4 hours",
		},
		{
			Name:        "zone",
			Label:       "placeholder filename",
			Type:        base.TypeString,
			Description: "default empty string",
			Default:     defaultPlaceholderName,
		},
		{
			Name:  "bool_1",
			Label: "S3ForcePathStyle",
			Type:  base.TypeBool,
		},
		{
			Name:    "internal_type",
			Label:   "ListObject Version",
			Type:    base.TypeSelect,
			Values:  "v1,v2",
			Default: "v1",
		},
	}
}

func (driver S3) Save(account *model.Account, old *model.Account) error {
	if account == nil {
		return nil
	}
	if account.Limit == 0 {
		account.Limit = 4
	}
	client, err := driver.NewSession(account)
	if err != nil {
		account.Status = err.Error()
	} else {
		sessionsMap[account.Name] = client
		account.Status = "work"
	}
	_ = model.SaveAccount(account)
	return err
}

func (driver S3) File(path string, account *model.Account) (*model.File, error) {
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

func (driver S3) Files(path string, account *model.Account) ([]model.File, error) {
	path = utils.ParsePath(path)
	var files []model.File
	cache, err := base.GetCache(path, account)
	if err == nil {
		files, _ = cache.([]model.File)
	} else {
		if account.InternalType == "v2" {
			files, err = driver.ListV2(path, account)
		} else {
			files, err = driver.List(path, account)
		}
		if err == nil && len(files) > 0 {
			_ = base.SetCache(path, files, account)
		}
	}
	return files, err
}

func (driver S3) Link(args base.Args, account *model.Account) (*base.Link, error) {
	client, err := driver.GetClient(account, true)
	if err != nil {
		return nil, err
	}
	path := driver.GetKey(args.Path, account, false)
	disposition := fmt.Sprintf(`attachment;filename="%s"`, url.QueryEscape(utils.Base(path)))
	input := &s3.GetObjectInput{
		Bucket: &account.Bucket,
		Key:    &path,
		//ResponseContentDisposition: &disposition,
	}
	if account.CustomHost == "" {
		input.ResponseContentDisposition = &disposition
	}
	req, _ := client.GetObjectRequest(input)
	var link string
	if account.CustomHost != "" {
		err = req.Build()
		link = req.HTTPRequest.URL.String()
	} else {
		link, err = req.Presign(time.Hour * time.Duration(account.Limit))
	}
	if err != nil {
		return nil, err
	}
	return &base.Link{
		Url: link,
	}, nil
}

func (driver S3) Path(path string, account *model.Account) (*model.File, []model.File, error) {
	path = utils.ParsePath(path)
	log.Debugf("s3 path: %s", path)
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

//func (driver S3) Proxy(r *http.Request, account *model.Account) {
//
//}

func (driver S3) Preview(path string, account *model.Account) (interface{}, error) {
	return nil, base.ErrNotSupport
}

func (driver S3) MakeDir(path string, account *model.Account) error {
	// not support, generate a placeholder file
	_, err := driver.File(path, account)
	// exist
	if err != base.ErrPathNotFound {
		return nil
	}
	return driver.Upload(&model.FileStream{
		File:       ioutil.NopCloser(bytes.NewReader([]byte{})),
		Size:       0,
		ParentPath: path,
		Name:       getPlaceholderName(account.Zone),
		MIMEType:   "application/octet-stream",
	}, account)
}

func (driver S3) Move(src string, dst string, account *model.Account) error {
	err := driver.Copy(src, dst, account)
	if err != nil {
		return err
	}
	return driver.Delete(src, account)
}

func (driver S3) Rename(src string, dst string, account *model.Account) error {
	return driver.Move(src, dst, account)
}

func (driver S3) Copy(src string, dst string, account *model.Account) error {
	client, err := driver.GetClient(account, false)
	if err != nil {
		return err
	}
	srcFile, err := driver.File(src, account)
	if err != nil {
		return err
	}
	srcKey := driver.GetKey(src, account, srcFile.IsDir())
	dstKey := driver.GetKey(dst, account, srcFile.IsDir())
	input := &s3.CopyObjectInput{
		Bucket:     &account.Bucket,
		CopySource: &srcKey,
		Key:        &dstKey,
	}
	_, err = client.CopyObject(input)
	return err
}

func (driver S3) Delete(path string, account *model.Account) error {
	client, err := driver.GetClient(account, false)
	if err != nil {
		return err
	}
	file, err := driver.File(path, account)
	if err != nil {
		return err
	}
	key := driver.GetKey(path, account, file.IsDir())
	input := &s3.DeleteObjectInput{
		Bucket: &account.Bucket,
		Key:    &key,
	}
	_, err = client.DeleteObject(input)
	return err
}

func (driver S3) Upload(file *model.FileStream, account *model.Account) error {
	if file == nil {
		return base.ErrEmptyFile
	}
	s, ok := sessionsMap[account.Name]
	if !ok {
		return fmt.Errorf("can't find [%s] session", account.Name)
	}
	uploader := s3manager.NewUploader(s)
	key := driver.GetKey(utils.Join(file.ParentPath, file.GetFileName()), account, false)
	log.Debugln("key:", key)
	input := &s3manager.UploadInput{
		Bucket: &account.Bucket,
		Key:    &key,
		Body:   file,
	}
	_, err := uploader.Upload(input)
	return err
}

var _ base.Driver = (*S3)(nil)
