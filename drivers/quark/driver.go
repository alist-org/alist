package quark

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Quark struct{}

func (driver Quark) Config() base.DriverConfig {
	return base.DriverConfig{
		Name:      "Quark",
		OnlyProxy: true,
	}
}

func (driver Quark) Items() []base.Item {
	return []base.Item{
		{
			Name:        "access_token",
			Label:       "Cookie",
			Type:        base.TypeString,
			Required:    true,
			Description: "Unknown expiration time",
		},
		{
			Name:     "root_folder",
			Label:    "root folder file_id",
			Type:     base.TypeString,
			Required: true,
			Default:  "0",
		},
		{
			Name:     "order_by",
			Label:    "order_by",
			Type:     base.TypeSelect,
			Values:   "file_type,file_name,updated_at",
			Required: true,
			Default:  "file_name",
		},
		{
			Name:     "order_direction",
			Label:    "order_direction",
			Type:     base.TypeSelect,
			Values:   "asc,desc",
			Required: true,
			Default:  "asc",
		},
	}
}

func (driver Quark) Save(account *model.Account, old *model.Account) error {
	if account == nil {
		return nil
	}
	_, err := driver.Get("/config", nil, nil, account)
	if err == nil {
		account.Status = "work"
	} else {
		account.Status = err.Error()
	}
	_ = model.SaveAccount(account)
	return err
}

func (driver Quark) File(path string, account *model.Account) (*model.File, error) {
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

func (driver Quark) Files(path string, account *model.Account) ([]model.File, error) {
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

func (driver Quark) Link(args base.Args, account *model.Account) (*base.Link, error) {
	path := args.Path
	file, err := driver.File(path, account)
	if err != nil {
		return nil, err
	}
	data := base.Json{
		"fids": []string{file.Id},
	}
	var resp DownResp
	_, err = driver.Post("/file/download", data, &resp, account)
	if err != nil {
		return nil, err
	}
	return &base.Link{
		Url: resp.Data[0].DownloadUrl,
		Headers: []base.Header{
			{Name: "Cookie", Value: account.AccessToken},
			{Name: "Referer", Value: "https://pan.quark.cn"},
		},
	}, nil
}

func (driver Quark) Path(path string, account *model.Account) (*model.File, []model.File, error) {
	path = utils.ParsePath(path)
	log.Debugf("quark path: %s", path)
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

func (driver Quark) Preview(path string, account *model.Account) (interface{}, error) {
	return nil, base.ErrNotSupport
}

func (driver Quark) MakeDir(path string, account *model.Account) error {
	parentFile, err := driver.File(utils.Dir(path), account)
	if err != nil {
		return err
	}
	data := base.Json{
		"dir_init_lock": false,
		"dir_path":      "",
		"file_name":     utils.Base(path),
		"pdir_fid":      parentFile.Id,
	}
	_, err = driver.Post("/file", data, nil, account)
	return err
}

func (driver Quark) Move(src string, dst string, account *model.Account) error {
	srcFile, err := driver.File(src, account)
	if err != nil {
		return err
	}
	dstParentFile, err := driver.File(utils.Dir(dst), account)
	if err != nil {
		return err
	}
	data := base.Json{
		"action_type":  1,
		"exclude_fids": []string{},
		"filelist":     []string{srcFile.Id},
		"to_pdir_fid":  dstParentFile.Id,
	}
	_, err = driver.Post("/file/move", data, nil, account)
	return err
}

func (driver Quark) Rename(src string, dst string, account *model.Account) error {
	srcFile, err := driver.File(src, account)
	if err != nil {
		return err
	}
	data := base.Json{
		"fid":       srcFile.Id,
		"file_name": utils.Base(dst),
	}
	_, err = driver.Post("/file/rename", data, nil, account)
	return err
}

func (driver Quark) Copy(src string, dst string, account *model.Account) error {
	return base.ErrNotSupport
}

func (driver Quark) Delete(path string, account *model.Account) error {
	srcFile, err := driver.File(path, account)
	if err != nil {
		return err
	}
	data := base.Json{
		"action_type":  1,
		"exclude_fids": []string{},
		"filelist":     []string{srcFile.Id},
	}
	_, err = driver.Post("/file/delete", data, nil, account)
	return err
}

func (driver Quark) Upload(file *model.FileStream, account *model.Account) error {
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
	defer func() {
		_ = tempFile.Close()
		_ = os.Remove(tempFile.Name())
	}()
	_, err = io.Copy(tempFile, file)
	if err != nil {
		return err
	}
	_, err = tempFile.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	m := md5.New()
	_, err = io.Copy(m, tempFile)
	if err != nil {
		return err
	}
	_, err = tempFile.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	md5Str := hex.EncodeToString(m.Sum(nil))
	s := sha1.New()
	_, err = io.Copy(s, tempFile)
	if err != nil {
		return err
	}
	_, err = tempFile.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	sha1Str := hex.EncodeToString(s.Sum(nil))
	// pre
	pre, err := driver.UpPre(file, parentFile.Id, account)
	if err != nil {
		return err
	}
	log.Debugln("hash: ", md5Str, sha1Str)
	// hash
	finish, err := driver.UpHash(md5Str, sha1Str, pre.Data.TaskId, account)
	if err != nil {
		return err
	}
	if finish {
		return nil
	}
	// part up
	partSize := pre.Metadata.PartSize
	var bytes []byte
	md5s := make([]string, 0)
	defaultBytes := make([]byte, partSize)
	left := int64(file.GetSize())
	partNumber := 1
	for left > 0 {
		if left > int64(partSize) {
			bytes = defaultBytes
		} else {
			bytes = make([]byte, left)
		}
		_, err := io.ReadFull(tempFile, bytes)
		if err != nil {
			return err
		}
		left -= int64(partSize)
		log.Debugf("left: %d", left)
		m, err := driver.UpPart(pre, file.GetMIMEType(), partNumber, bytes, account)
		//m, err := driver.UpPart(pre, file.GetMIMEType(), partNumber, bytes, account, md5Str, sha1Str)
		if err != nil {
			return err
		}
		if m == "finish" {
			return nil
		}
		md5s = append(md5s, m)
		partNumber++
	}
	err = driver.UpCommit(pre, md5s, account)
	if err != nil {
		return err
	}
	return driver.UpFinish(pre, account)
}

var _ base.Driver = (*Quark)(nil)
