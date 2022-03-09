package xunlei

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"

	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	log "github.com/sirupsen/logrus"
)

type XunLeiCloud struct{}

func init() {
	base.RegisterDriver(new(XunLeiCloud))
}

func (driver XunLeiCloud) Config() base.DriverConfig {
	return base.DriverConfig{
		Name:      "XunLeiCloud",
		LocalSort: true,
	}
}

func (driver XunLeiCloud) Items() []base.Item {
	return []base.Item{
		{
			Name:        "username",
			Label:       "username",
			Type:        base.TypeString,
			Required:    true,
			Description: "account username/phone number",
		},
		{
			Name:        "password",
			Label:       "password",
			Type:        base.TypeString,
			Required:    true,
			Description: "account password",
		},
		{
			Name:     "root_folder",
			Label:    "root folder file_id",
			Type:     base.TypeString,
			Required: true,
		},
	}
}

func (driver XunLeiCloud) Save(account *model.Account, old *model.Account) error {
	if account == nil {
		return nil
	}
	state := GetState(account)
	if state.isTokensExpires() {
		return state.Login(account)
	}
	account.Status = "work"
	model.SaveAccount(account)
	return nil
}

func (driver XunLeiCloud) File(path string, account *model.Account) (*model.File, error) {
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

func (driver XunLeiCloud) Files(path string, account *model.Account) ([]model.File, error) {
	cache, err := base.GetCache(path, account)
	if err == nil {
		files, _ := cache.([]model.File)
		return files, nil
	}
	file, err := driver.File(utils.ParsePath(path), account)
	if err != nil {
		return nil, err
	}

	files := make([]model.File, 0)
	for {
		var fileList FileList
		u := fmt.Sprintf("https://api-pan.xunlei.com/drive/v1/files?parent_id=%s&page_token=%s&with_audit=true&filters=%s", file.Id, fileList.NextPageToken, url.QueryEscape(`{"phase": {"eq": "PHASE_TYPE_COMPLETE"}, "trashed":{"eq":false}}`))
		if err = GetState(account).Request("GET", u, nil, &fileList, account); err != nil {
			return nil, err
		}
		for _, file := range fileList.Files {
			if file.Kind == FOLDER || (file.Kind == FILE && file.Audit.Status == "STATUS_OK") {
				files = append(files, *driver.formatFile(&file))
			}
		}
		if fileList.NextPageToken == "" {
			break
		}
	}
	if len(files) > 0 {
		_ = base.SetCache(path, files, account)
	}
	return files, nil
}

func (driver XunLeiCloud) formatFile(file *Files) *model.File {
	size, _ := strconv.ParseInt(file.Size, 10, 64)
	tp := conf.FOLDER
	if file.Kind == FILE {
		tp = utils.GetFileType(file.FileExtension)
	}
	return &model.File{
		Id:        file.ID,
		Name:      file.Name,
		Size:      size,
		Type:      tp,
		Driver:    driver.Config().Name,
		UpdatedAt: file.CreatedTime,
		Thumbnail: file.ThumbnailLink,
	}
}

func (driver XunLeiCloud) Link(args base.Args, account *model.Account) (*base.Link, error) {
	file, err := driver.File(utils.ParsePath(args.Path), account)
	if err != nil {
		return nil, err
	}
	if file.Type == conf.FOLDER {
		return nil, base.ErrNotFile
	}
	var lFile Files
	if err = GetState(account).Request("GET", fmt.Sprintf("https://api-pan.xunlei.com/drive/v1/files/%s?&with_audit=true", file.Id), nil, &lFile, account); err != nil {
		return nil, err
	}
	return &base.Link{
		Headers: []base.Header{
			{Name: "User-Agent", Value: base.UserAgent},
		},
		Url: lFile.WebContentLink,
	}, nil
}

func (driver XunLeiCloud) Path(path string, account *model.Account) (*model.File, []model.File, error) {
	path = utils.ParsePath(path)
	log.Debugf("xunlei path: %s", path)
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

func (driver XunLeiCloud) Preview(path string, account *model.Account) (interface{}, error) {
	return nil, base.ErrNotSupport
}

func (driver XunLeiCloud) MakeDir(path string, account *model.Account) error {
	dir, name := filepath.Split(path)
	parentFile, err := driver.File(dir, account)
	if err != nil {
		return err
	}
	if !parentFile.IsDir() {
		return base.ErrNotFolder
	}
	return GetState(account).Request("POST", "https://api-pan.xunlei.com/drive/v1/files", &base.Json{"kind": FOLDER, "name": name, "parent_id": parentFile.Id}, nil, account)
}

func (driver XunLeiCloud) Move(src string, dst string, account *model.Account) error {
	srcFile, err := driver.File(src, account)
	if err != nil {
		return err
	}

	dstDirFile, err := driver.File(filepath.Dir(dst), account)
	if err != nil {
		return err
	}
	return GetState(account).Request("POST", "https://api-pan.xunlei.com/drive/v1/files:batchMove", &base.Json{"to": base.Json{"parent_id": dstDirFile.Id}, "ids": []string{srcFile.Id}}, nil, account)
}

func (driver XunLeiCloud) Copy(src string, dst string, account *model.Account) error {
	srcFile, err := driver.File(src, account)
	if err != nil {
		return err
	}

	dstDirFile, err := driver.File(filepath.Dir(dst), account)
	if err != nil {
		return err
	}
	return GetState(account).Request("POST", "https://api-pan.xunlei.com/drive/v1/files:batchCopy", &base.Json{"to": base.Json{"parent_id": dstDirFile.Id}, "ids": []string{srcFile.Id}}, nil, account)
}

func (driver XunLeiCloud) Delete(path string, account *model.Account) error {
	srcFile, err := driver.File(path, account)
	if err != nil {
		return err
	}
	return GetState(account).Request("PATCH", fmt.Sprintf("https://api-pan.xunlei.com/drive/v1/files/%s/trash", srcFile.Id), &base.Json{}, nil, account)
}

func (driver XunLeiCloud) Upload(file *model.FileStream, account *model.Account) error {
	if file == nil {
		return base.ErrEmptyFile
	}

	parentFile, err := driver.File(file.ParentPath, account)
	if err != nil {
		return err
	}

	tempFile, err := ioutil.TempFile(conf.Conf.TempDir, "file-*")
	if err != nil {
		return err
	}

	defer tempFile.Close()
	defer os.Remove(tempFile.Name())

	gcid, err := getGcid(io.TeeReader(file, tempFile), int64(file.Size))
	if err != nil {
		return err
	}

	var rep UploadTaskResponse
	err = GetState(account).Request("POST", "https://api-pan.xunlei.com/drive/v1/files", &base.Json{
		"kind":        FILE,
		"parent_id":   parentFile.Id,
		"name":        file.Name,
		"size":        fmt.Sprint(file.Size),
		"hash":        gcid,
		"upload_type": UPLOAD_TYPE_RESUMABLE,
	}, &rep, account)
	if err != nil {
		return err
	}

	param := rep.Resumable.Params
	client, err := oss.New(param.Endpoint, param.AccessKeyID, param.AccessKeySecret, oss.SecurityToken(param.SecurityToken), oss.EnableMD5(true), oss.HTTPClient(xunleiClient.GetClient()))
	if err != nil {
		return err
	}
	bucket, err := client.Bucket(param.Bucket)
	if err != nil {
		return err
	}
	return bucket.UploadFile(param.Key, tempFile.Name(), 4*1024*1024, oss.Routines(3), oss.Checkpoint(true, ""), oss.Expires(param.Expiration))
}

func (driver XunLeiCloud) Rename(src string, dst string, account *model.Account) error {
	_, dstName := filepath.Split(dst)
	srcFile, err := driver.File(src, account)
	if err != nil {
		return err
	}

	return GetState(account).Request("PATCH", fmt.Sprintf("https://api-pan.xunlei.com/drive/v1/files/%s", srcFile.Id), &base.Json{"name": dstName}, nil, account)
}

var _ base.Driver = (*XunLeiCloud)(nil)
