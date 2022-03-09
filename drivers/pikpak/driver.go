package pikpak

import (
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	"path/filepath"
	"strings"
)

type PikPak struct{}

func (driver PikPak) Config() base.DriverConfig {
	return base.DriverConfig{
		Name:      "PikPak",
		ApiProxy:  true,
		LocalSort: true,
	}
}

func (driver PikPak) Items() []base.Item {
	return []base.Item{
		{
			Name:     "username",
			Label:    "username",
			Type:     base.TypeString,
			Required: true,
		},
		{
			Name:     "password",
			Label:    "password",
			Type:     base.TypeString,
			Required: true,
		},
		{
			Name:     "root_folder",
			Label:    "root folder id",
			Type:     base.TypeString,
			Required: false,
		},
	}
}

func (driver PikPak) Save(account *model.Account, old *model.Account) error {
	if account == nil {
		return nil
	}
	err := driver.Login(account)
	return err
}

func (driver PikPak) File(path string, account *model.Account) (*model.File, error) {
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

func (driver PikPak) Files(path string, account *model.Account) ([]model.File, error) {
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
		rawFiles, err := driver.GetFiles(file.Id, account)
		if err != nil {
			return nil, err
		}
		files = make([]model.File, 0)
		for _, file := range rawFiles {
			files = append(files, *driver.FormatFile(&file))
		}
		if len(files) > 0 {
			_ = base.SetCache(path, files, account)
		}
	}
	return files, nil
}

func (driver PikPak) Link(args base.Args, account *model.Account) (*base.Link, error) {
	file, err := driver.File(args.Path, account)
	if err != nil {
		return nil, err
	}
	var resp File
	_, err = driver.Request(fmt.Sprintf("https://api-drive.mypikpak.com/drive/v1/files/%s?_magic=2021&thumbnail_size=SIZE_LARGE", file.Id),
		base.Get, nil, nil, &resp, account)
	if err != nil {
		return nil, err
	}
	link := base.Link{
		Url: resp.WebContentLink,
	}
	if len(resp.Medias) > 0 && resp.Medias[0].Link.Url != "" {
		log.Debugln("use media link")
		link.Url = resp.Medias[0].Link.Url
	}
	return &link, nil
}

func (driver PikPak) Path(path string, account *model.Account) (*model.File, []model.File, error) {
	path = utils.ParsePath(path)
	log.Debugf("pikpak path: %s", path)
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

//func (driver PikPak) Proxy(r *http.Request, account *model.Account) {
//
//}

func (driver PikPak) Preview(path string, account *model.Account) (interface{}, error) {
	return nil, base.ErrNotSupport
}

func (driver PikPak) MakeDir(path string, account *model.Account) error {
	path = utils.ParsePath(path)
	dir, name := filepath.Split(path)
	parentFile, err := driver.File(dir, account)
	if err != nil {
		return err
	}
	if !parentFile.IsDir() {
		return base.ErrNotFolder
	}
	_, err = driver.Request("https://api-drive.mypikpak.com/drive/v1/files", base.Post, nil, &base.Json{
		"kind":      "drive#folder",
		"parent_id": parentFile.Id,
		"name":      name,
	}, nil, account)
	return err
}

func (driver PikPak) Move(src string, dst string, account *model.Account) error {
	dstDir, _ := filepath.Split(dst)
	srcFile, err := driver.File(src, account)
	if err != nil {
		return err
	}
	dstDirFile, err := driver.File(dstDir, account)
	if err != nil {
		return err
	}
	_, err = driver.Request("https://api-drive.mypikpak.com/drive/v1/files:batchMove", base.Post, nil, &base.Json{
		"ids": []string{srcFile.Id},
		"to": base.Json{
			"parent_id": dstDirFile.Id,
		},
	}, nil, account)
	return err
}

func (driver PikPak) Rename(src string, dst string, account *model.Account) error {
	_, dstName := filepath.Split(dst)
	srcFile, err := driver.File(src, account)
	if err != nil {
		return err
	}
	_, err = driver.Request("https://api-drive.mypikpak.com/drive/v1/files/"+srcFile.Id, base.Patch, nil, &base.Json{
		"name": dstName,
	}, nil, account)
	return err
}

func (driver PikPak) Copy(src string, dst string, account *model.Account) error {
	srcFile, err := driver.File(src, account)
	if err != nil {
		return err
	}
	dstDirFile, err := driver.File(utils.Dir(dst), account)
	if err != nil {
		return err
	}
	_, err = driver.Request("https://api-drive.mypikpak.com/drive/v1/files:batchCopy", base.Post, nil, &base.Json{
		"ids": []string{srcFile.Id},
		"to": base.Json{
			"parent_id": dstDirFile.Id,
		},
	}, nil, account)
	return err
}

func (driver PikPak) Delete(path string, account *model.Account) error {
	file, err := driver.File(path, account)
	if err != nil {
		return err
	}
	_, err = driver.Request("https://api-drive.mypikpak.com/drive/v1/files:batchTrash", base.Post, nil, &base.Json{
		"ids": []string{file.Id},
	}, nil, account)
	return err
}

func (driver PikPak) Upload(file *model.FileStream, account *model.Account) error {
	if file == nil {
		return base.ErrEmptyFile
	}
	parentFile, err := driver.File(file.ParentPath, account)
	if err != nil {
		return err
	}
	data := base.Json{
		"kind":        "drive#file",
		"name":        file.GetFileName(),
		"size":        file.GetSize(),
		"hash":        "1CF254FBC456E1B012CD45C546636AA62CF8350E",
		"upload_type": "UPLOAD_TYPE_RESUMABLE",
		"objProvider": base.Json{"provider": "UPLOAD_TYPE_UNKNOWN"},
		"parent_id":   parentFile.Id,
	}
	res, err := driver.Request("https://api-drive.mypikpak.com/drive/v1/files", base.Post, nil, &data, nil, account)
	if err != nil {
		return err
	}
	params := jsoniter.Get(res, "resumable").Get("params")
	endpoint := params.Get("endpoint").ToString()
	endpointS := strings.Split(endpoint, ".")
	endpoint = strings.Join(endpointS[1:], ".")
	accessKeyId := params.Get("access_key_id").ToString()
	accessKeySecret := params.Get("access_key_secret").ToString()
	securityToken := params.Get("security_token").ToString()
	key := params.Get("key").ToString()
	bucket := params.Get("bucket").ToString()
	cfg := &aws.Config{
		Credentials: credentials.NewStaticCredentials(accessKeyId, accessKeySecret, securityToken),
		Region:      aws.String("pikpak"),
		Endpoint:    &endpoint,
	}
	s, err := session.NewSession(cfg)
	if err != nil {
		return err
	}
	uploader := s3manager.NewUploader(s)
	input := &s3manager.UploadInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   file,
	}
	_, err = uploader.Upload(input)
	return err
}

// use aliyun-oss-sdk
//func (driver PikPak) Upload(file *model.FileStream, account *model.Account) error {
//	if file == nil {
//		return base.ErrEmptyFile
//	}
//	parentFile, err := driver.File(file.ParentPath, account)
//	if err != nil {
//		return err
//	}
//	data := base.Json{
//		"kind":        "drive#file",
//		"name":        file.GetFileName(),
//		"size":        file.GetSize(),
//		"hash":        "1CF254FBC456E1B012CD45C546636AA62CF8350E",
//		"upload_type": "UPLOAD_TYPE_RESUMABLE",
//		"objProvider": base.Json{"provider": "UPLOAD_TYPE_UNKNOWN"},
//		"parent_id":   parentFile.Id,
//	}
//	res, err := driver.Request("https://api-drive.mypikpak.com/drive/v1/files", base.Post, nil, &data, nil, account)
//	if err != nil {
//		return err
//	}
//	params := jsoniter.Get(res, "resumable").Get("params")
//	endpoint := params.Get("endpoint").ToString()
//	endpointS := strings.Split(endpoint, ".")
//	endpoint = strings.Join(endpointS[1:], ".")
//	accessKeyId := params.Get("access_key_id").ToString()
//	accessKeySecret := params.Get("access_key_secret").ToString()
//	securityToken := params.Get("security_token").ToString()
//	client, err := oss.New("https://"+endpoint, accessKeyId,
//		accessKeySecret, oss.SecurityToken(securityToken))
//	if err != nil {
//		return err
//	}
//	bucket, err := client.Bucket(params.Get("bucket").ToString())
//	if err != nil {
//		return err
//	}
//	signedURL, err := bucket.SignURL(params.Get("key").ToString(), oss.HTTPPut, 60)
//	if err != nil {
//		return err
//	}
//	err = bucket.PutObjectWithURL(signedURL, file)
//	return err
//}

var _ base.Driver = (*PikPak)(nil)
