package alidrive

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"io"
	"math"
	"net/http"
	"path/filepath"
)

type AliDrive struct{}

func (driver AliDrive) Config() base.DriverConfig {
	return base.DriverConfig{
		Name: "AliDrive",
	}
}

func (driver AliDrive) Items() []base.Item {
	return []base.Item{
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
			Name:     "order_by",
			Label:    "order_by",
			Type:     base.TypeSelect,
			Values:   "name,size,updated_at,created_at",
			Required: false,
		},
		{
			Name:     "order_direction",
			Label:    "order_direction",
			Type:     base.TypeSelect,
			Values:   "ASC,DESC",
			Required: false,
		},
		{
			Name:        "limit",
			Label:       "limit",
			Type:        base.TypeNumber,
			Required:    false,
			Description: ">0 and <=200",
		},
	}
}

func (driver AliDrive) Save(account *model.Account, old *model.Account) error {
	if old != nil {
		conf.Cron.Remove(cron.EntryID(old.CronId))
	}
	if account == nil {
		return nil
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
	var resp base.Json
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
	return nil, base.ErrPathNotFound
}

func (driver AliDrive) Files(path string, account *model.Account) ([]model.File, error) {
	path = utils.ParsePath(path)
	var rawFiles []AliFile
	cache, err := base.GetCache(path, account)
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
			_ = base.SetCache(path, rawFiles, account)
		}
	}
	files := make([]model.File, 0)
	for _, file := range rawFiles {
		files = append(files, *driver.FormatFile(&file))
	}
	return files, nil
}

func (driver AliDrive) Link(args base.Args, account *model.Account) (*base.Link, error) {
	file, err := driver.File(args.Path, account)
	if err != nil {
		return nil, err
	}
	var resp base.Json
	var e AliRespError
	_, err = aliClient.R().SetResult(&resp).
		SetError(&e).
		SetHeader("authorization", "Bearer\t"+account.AccessToken).
		SetBody(base.Json{
			"drive_id":   account.DriveId,
			"file_id":    file.Id,
			"expire_sec": 14400,
		}).Post("https://api.aliyundrive.com/v2/file/get_download_url")
	if err != nil {
		return nil, err
	}
	if e.Code != "" {
		if e.Code == "AccessTokenInvalid" {
			err = driver.RefreshToken(account)
			if err != nil {
				return nil, err
			} else {
				_ = model.SaveAccount(account)
				return driver.Link(args, account)
			}
		}
		return nil, fmt.Errorf("%s", e.Message)
	}
	return &base.Link{
		Headers: []base.Header{
			{
				Name:  "Referer",
				Value: "https://www.aliyundrive.com/",
			},
		},
		Url: resp["url"].(string),
	}, nil
}

func (driver AliDrive) Path(path string, account *model.Account) (*model.File, []model.File, error) {
	path = utils.ParsePath(path)
	log.Debugf("ali path: %s", path)
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

//func (driver AliDrive) Proxy(r *http.Request, account *model.Account) {
//	r.Header.Del("Origin")
//	r.Header.Set("Referer", "https://www.aliyundrive.com/")
//}

func (driver AliDrive) Preview(path string, account *model.Account) (interface{}, error) {
	file, err := driver.GetFile(path, account)
	if err != nil {
		return nil, err
	}
	// office
	var resp base.Json
	var e AliRespError
	var url string
	req := base.Json{
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
		return nil, base.ErrNotSupport
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

func (driver AliDrive) MakeDir(path string, account *model.Account) error {
	dir, name := filepath.Split(path)
	parentFile, err := driver.File(dir, account)
	if err != nil {
		return err
	}
	if !parentFile.IsDir() {
		return base.ErrNotFolder
	}
	var resp base.Json
	var e AliRespError
	_, err = aliClient.R().SetResult(&resp).SetError(&e).
		SetHeader("authorization", "Bearer\t"+account.AccessToken).
		SetBody(base.Json{
			"check_name_mode": "refuse",
			"drive_id":        account.DriveId,
			"name":            name,
			"parent_file_id":  parentFile.Id,
			"type":            "folder",
		}).Post("https://api.aliyundrive.com/adrive/v2/file/createWithFolders")
	if e.Code != "" {
		if e.Code == "AccessTokenInvalid" {
			err = driver.RefreshToken(account)
			if err != nil {
				return err
			} else {
				_ = model.SaveAccount(account)
				return driver.MakeDir(path, account)
			}
		}
		return fmt.Errorf("%s", e.Message)
	}
	if resp["file_name"] == name {
		return nil
	}
	return fmt.Errorf("%+v", resp)
}

func (driver AliDrive) Move(src string, dst string, account *model.Account) error {
	dstDir, _ := filepath.Split(dst)
	srcFile, err := driver.File(src, account)
	if err != nil {
		return err
	}
	dstDirFile, err := driver.File(dstDir, account)
	if err != nil {
		return err
	}
	err = driver.batch(srcFile.Id, dstDirFile.Id, account)
	return err
}

func (driver AliDrive) Rename(src string, dst string, account *model.Account) error {
	_, dstName := filepath.Split(dst)
	srcFile, err := driver.File(src, account)
	if err != nil {
		return err
	}
	err = driver.rename(srcFile.Id, dstName, account)
	return err
}

func (driver AliDrive) Copy(src string, dst string, account *model.Account) error {
	return base.ErrNotSupport
}

func (driver AliDrive) Delete(path string, account *model.Account) error {
	file, err := driver.File(path, account)
	if err != nil {
		return err
	}
	var e AliRespError
	res, err := aliClient.R().SetError(&e).
		SetHeader("authorization", "Bearer\t"+account.AccessToken).
		SetBody(base.Json{
			"drive_id": account.DriveId,
			"file_id":  file.Id,
		}).Post("https://api.aliyundrive.com/v2/recyclebin/trash")
	if err != nil {
		return err
	}
	if e.Code != "" {
		if e.Code == "AccessTokenInvalid" {
			err = driver.RefreshToken(account)
			if err != nil {
				return err
			} else {
				_ = model.SaveAccount(account)
				return driver.Delete(path, account)
			}
		}
		return fmt.Errorf("%s", e.Message)
	}
	if res.StatusCode() == 204 {
		return nil
	}
	return errors.New(res.String())
}

type UploadResp struct {
	FileId       string `json:"file_id"`
	UploadId     string `json:"upload_id"`
	PartInfoList []struct {
		UploadUrl string `json:"upload_url"`
	} `json:"part_info_list"`
}

func (driver AliDrive) Upload(file *model.FileStream, account *model.Account) error {
	if file == nil {
		return base.ErrEmptyFile
	}
	const DEFAULT uint64 = 10485760
	var count = int64(math.Ceil(float64(file.GetSize()) / float64(DEFAULT)))
	var finish uint64 = 0
	parentFile, err := driver.File(file.ParentPath, account)
	if err != nil {
		return err
	}
	if !parentFile.IsDir() {
		return base.ErrNotFolder
	}
	var resp UploadResp
	var e AliRespError
	partInfoList := make([]base.Json, 0)
	var i int64
	for i = 0; i < count; i++ {
		partInfoList = append(partInfoList, base.Json{
			"part_number": i + 1,
		})
	}
	_, err = aliClient.R().SetResult(&resp).SetError(&e).
		SetHeader("authorization", "Bearer\t"+account.AccessToken).
		SetBody(base.Json{
			"check_name_mode": "auto_rename",
			// content_hash
			"content_hash_name": "none",
			"drive_id":          account.DriveId,
			"name":              file.GetFileName(),
			"parent_file_id":    parentFile.Id,
			"part_info_list":    partInfoList,
			//proof_code
			"proof_version": "v1",
			"size":          file.GetSize(),
			"type":          "file",
		}).Post("https://api.aliyundrive.com/adrive/v2/file/createWithFolders") // /v2/file/create_with_proof
	//log.Debugf("%+v\n%+v", resp, e)
	if e.Code != "" {
		if e.Code == "AccessTokenInvalid" {
			err = driver.RefreshToken(account)
			if err != nil {
				return err
			} else {
				_ = model.SaveAccount(account)
				return driver.Upload(file, account)
			}
		}
		return fmt.Errorf("%s", e.Message)
	}
	var byteSize uint64
	for i = 0; i < count; i++ {
		byteSize = file.GetSize() - finish
		if DEFAULT < byteSize {
			byteSize = DEFAULT
		}
		log.Debugf("%d,%d", byteSize, finish)
		byteData := make([]byte, byteSize)
		n, err := io.ReadFull(file, byteData)
		//n, err := file.Read(byteData)
		//byteData, err := io.ReadAll(file)
		//n := len(byteData)
		log.Debug(err, n)
		if err != nil {
			return err
		}

		finish += uint64(n)

		req, err := http.NewRequest("PUT", resp.PartInfoList[i].UploadUrl, bytes.NewBuffer(byteData))
		if err != nil {
			return err
		}
		res, err := base.HttpClient.Do(req)
		if err != nil {
			return err
		}
		log.Debugf("%+v", res)
		//res, err := base.BaseClient.R().
		//	SetHeader("Content-Type","").
		//	SetBody(byteData).Put(resp.PartInfoList[i].UploadUrl)
		//if err != nil {
		//	return err
		//}
		//log.Debugf("put to %s : %d,%s", resp.PartInfoList[i].UploadUrl, res.StatusCode(),res.String())
	}
	var resp2 base.Json
	_, err = aliClient.R().SetResult(&resp2).SetError(&e).
		SetHeader("authorization", "Bearer\t"+account.AccessToken).
		SetBody(base.Json{
			"drive_id":  account.DriveId,
			"file_id":   resp.FileId,
			"upload_id": resp.UploadId,
		}).Post("https://api.aliyundrive.com/v2/file/complete")
	if e.Code != "" {
		//if e.Code == "AccessTokenInvalid" {
		//	err = driver.RefreshToken(account)
		//	if err != nil {
		//		return err
		//	} else {
		//		_ = model.SaveAccount(account)
		//		return driver.Upload(file, account)
		//	}
		//}
		return fmt.Errorf("%s", e.Message)
	}
	if resp2["file_id"] == resp.FileId {
		return nil
	}
	return fmt.Errorf("%+v", resp2)
}

var _ base.Driver = (*AliDrive)(nil)
