package alidrive

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/big"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
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
		{
			Name:  "bool_1",
			Label: "fast upload",
			Type:  base.TypeBool,
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
		id := account.ID
		log.Debugf("ali account id: %d", id)
		newAccount, err := model.GetAccountById(id)
		log.Debugf("ali account: %+v", newAccount)
		if err != nil {
			return
		}
		err = driver.RefreshToken(newAccount)
		_ = model.SaveAccount(newAccount)
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
	err = driver.batch(srcFile.Id, dstDirFile.Id, "/file/move", account)
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
	dstDir, _ := filepath.Split(dst)
	srcFile, err := driver.File(src, account)
	if err != nil {
		return err
	}
	dstDirFile, err := driver.File(dstDir, account)
	if err != nil {
		return err
	}
	err = driver.batch(srcFile.Id, dstDirFile.Id, "/file/copy", account)
	return err
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
	if res.StatusCode() < 400 {
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

	RapidUpload bool `json:"rapid_upload"`
}

func (driver AliDrive) Upload(file *model.FileStream, account *model.Account) error {
	if file == nil {
		return base.ErrEmptyFile
	}

	parentFile, err := driver.File(file.ParentPath, account)
	if err != nil {
		return err
	}
	if !parentFile.IsDir() {
		return base.ErrNotFolder
	}

	const DEFAULT int64 = 10485760
	var count = int(math.Ceil(float64(file.GetSize()) / float64(DEFAULT)))

	partInfoList := make([]base.Json, 0, count)
	for i := 1; i <= count; i++ {
		partInfoList = append(partInfoList, base.Json{"part_number": i})
	}

	reqBody := base.Json{
		"check_name_mode": "overwrite",
		"drive_id":        account.DriveId,
		"name":            file.GetFileName(),
		"parent_file_id":  parentFile.Id,
		"part_info_list":  partInfoList,
		"size":            file.GetSize(),
		"type":            "file",
	}

	if account.Bool1 {
		buf := bytes.NewBuffer(make([]byte, 0, 1024))
		io.CopyN(buf, file, 1024)
		reqBody["pre_hash"] = utils.GetSHA1Encode(buf.String())
		// 把头部拼接回去
		file.File = struct {
			io.Reader
			io.Closer
		}{
			Reader: io.MultiReader(buf, file.File),
			Closer: file.File,
		}
	} else {
		reqBody["content_hash_name"] = "none"
		reqBody["proof_version"] = "v1"
	}

	var resp UploadResp
	var e AliRespError
	client := aliClient.R().SetResult(&resp).SetError(&e).SetHeader("authorization", "Bearer\t"+account.AccessToken).SetBody(reqBody)

	_, err = client.Post("https://api.aliyundrive.com/adrive/v2/file/createWithFolders")
	if err != nil {
		return err
	}
	if e.Code != "" && e.Code != "PreHashMatched" {
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

	if account.Bool1 && e.Code == "PreHashMatched" {
		tempFile, err := ioutil.TempFile(conf.Conf.TempDir, "file-*")
		if err != nil {
			return err
		}

		defer func() {
			_ = tempFile.Close()
			_ = os.Remove(tempFile.Name())
		}()

		delete(reqBody, "pre_hash")
		h := sha1.New()
		if _, err = io.Copy(io.MultiWriter(tempFile, h), file.File); err != nil {
			return err
		}
		reqBody["content_hash"] = hex.EncodeToString(h.Sum(nil))
		reqBody["content_hash_name"] = "sha1"
		reqBody["proof_version"] = "v1"

		/*
			js 隐性转换太坑不知道有没有bug
			var n = e.access_token，
			r = new BigNumber('0x'.concat(md5(n).slice(0, 16)))，
			i = new BigNumber(t.file.size)，
			o = i ? r.mod(i) : new gt.BigNumber(0);
			(t.file.slice(o.toNumber(), Math.min(o.plus(8).toNumber(), t.file.size)))
		*/
		buf := make([]byte, 8)
		r, _ := new(big.Int).SetString(utils.GetMD5Encode(account.AccessToken)[:16], 16)
		i := new(big.Int).SetUint64(file.Size)
		o := r.Mod(r, i)
		n, _ := io.NewSectionReader(tempFile, o.Int64(), 8).Read(buf[:8])
		reqBody["proof_code"] = base64.StdEncoding.EncodeToString(buf[:n])

		_, err = client.Post("https://api.aliyundrive.com/adrive/v2/file/createWithFolders")
		if err != nil {
			return err
		}
		if e.Code != "" && e.Code != "PreHashMatched" {
			return fmt.Errorf("%s", e.Message)
		}

		if resp.RapidUpload {
			return nil
		}

		// 秒传失败
		if _, err = tempFile.Seek(0, io.SeekStart); err != nil {
			return err
		}
		file.File = tempFile
	}

	for _, partInfo := range resp.PartInfoList {
		req, err := http.NewRequest("PUT", partInfo.UploadUrl, io.LimitReader(file.File, DEFAULT))
		if err != nil {
			return err
		}
		res, err := base.HttpClient.Do(req)
		if err != nil {
			return err
		}
		log.Debugf("%+v", res)
		res.Body.Close()
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
	if err != nil {
		return err
	}
	if e.Code != "" && e.Code != "PreHashMatched" {
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
