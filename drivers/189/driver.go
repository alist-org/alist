package _89

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	"io"
	"math"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

type Cloud189 struct{}

func (driver Cloud189) Config() base.DriverConfig {
	return base.DriverConfig{
		Name: "189Cloud",
	}
}

func (driver Cloud189) Items() []base.Item {
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
		{
			Name:     "order_by",
			Label:    "order_by",
			Type:     base.TypeSelect,
			Values:   "name,size,lastOpTime,createdDate",
			Required: true,
		},
		{
			Name:     "order_direction",
			Label:    "desc",
			Type:     base.TypeSelect,
			Values:   "true,false",
			Required: true,
		},
	}
}

func (driver Cloud189) Save(account *model.Account, old *model.Account) error {
	if old != nil && old.Name != account.Name {
		delete(client189Map, old.Name)
	}
	if err := driver.Login(account); err != nil {
		account.Status = err.Error()
		_ = model.SaveAccount(account)
		return err
	}
	account.Status = "work"
	err := model.SaveAccount(account)
	if err != nil {
		return err
	}
	return nil
}

func (driver Cloud189) File(path string, account *model.Account) (*model.File, error) {
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

func (driver Cloud189) Files(path string, account *model.Account) ([]model.File, error) {
	path = utils.ParsePath(path)
	var rawFiles []Cloud189File
	cache, err := base.GetCache(path, account)
	if err == nil {
		rawFiles, _ = cache.([]Cloud189File)
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

func (driver Cloud189) Link(args base.Args, account *model.Account) (*base.Link, error) {
	file, err := driver.File(utils.ParsePath(args.Path), account)
	if err != nil {
		return nil, err
	}
	if file.Type == conf.FOLDER {
		return nil, base.ErrNotFile
	}
	client, ok := client189Map[account.Name]
	if !ok {
		return nil, fmt.Errorf("can't find [%s] client", account.Name)
	}
	var e Cloud189Error
	var resp Cloud189Down
	_, err = client.R().SetResult(&resp).SetError(&e).
		SetHeader("Accept", "application/json;charset=UTF-8").
		SetQueryParams(map[string]string{
			"noCache": random(),
			"fileId":  file.Id,
		}).Get("https://cloud.189.cn/api/open/file/getFileDownloadUrl.action")
	if err != nil {
		return nil, err
	}
	if e.ErrorCode != "" {
		if e.ErrorCode == "InvalidSessionKey" {
			err = driver.Login(account)
			if err != nil {
				return nil, err
			}
			return driver.Link(args, account)
		}
	}
	if resp.ResCode != 0 {
		return nil, fmt.Errorf(resp.ResMessage)
	}
	res, err := base.NoRedirectClient.R().Get(resp.FileDownloadUrl)
	if err != nil {
		return nil, err
	}
	link := base.Link{}
	if res.StatusCode() == 302 {
		link.Url = res.Header().Get("location")
	} else {
		link.Url = resp.FileDownloadUrl
	}
	return &link, nil
}

func (driver Cloud189) Path(path string, account *model.Account) (*model.File, []model.File, error) {
	path = utils.ParsePath(path)
	log.Debugf("189 path: %s", path)
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

func (driver Cloud189) Proxy(ctx *gin.Context, account *model.Account) {
	ctx.Request.Header.Del("Origin")
}

func (driver Cloud189) Preview(path string, account *model.Account) (interface{}, error) {
	return nil, base.ErrNotSupport
}

func (driver Cloud189) MakeDir(path string, account *model.Account) error {
	dir, name := filepath.Split(path)
	parent, err := driver.File(dir, account)
	if err != nil {
		return err
	}
	if !parent.IsDir() {
		return base.ErrNotFolder
	}
	form := map[string]string{
		"parentFolderId": parent.Id,
		"folderName":     name,
	}
	_, err = driver.Request("https://cloud.189.cn/api/open/file/createFolder.action", "POST", form, nil, account)
	return err
}

func (driver Cloud189) Move(src string, dst string, account *model.Account) error {
	dstDir, dstName := filepath.Split(dst)
	srcFile, err := driver.File(src, account)
	if err != nil {
		return err
	}

	dstDirFile, err := driver.File(dstDir, account)
	if err != nil {
		return err
	}
	isFolder := 0
	if srcFile.IsDir() {
		isFolder = 1
	}
	taskInfos := []base.Json{
		{
			"fileId":   srcFile.Id,
			"fileName": dstName,
			"isFolder": isFolder,
		},
	}
	taskInfosBytes, err := json.Marshal(taskInfos)
	if err != nil {
		return err
	}
	form := map[string]string{
		"type":           "MOVE",
		"targetFolderId": dstDirFile.Id,
		"taskInfos":      string(taskInfosBytes),
	}
	_, err = driver.Request("https://cloud.189.cn/api/open/batch/createBatchTask.action", "POST", form, nil, account)
	return err
}

func (driver Cloud189) Rename(src string, dst string, account *model.Account) error {
	_, dstName := filepath.Split(dst)
	srcFile, err := driver.File(src, account)
	if err != nil {
		return err
	}
	url := "https://cloud.189.cn/api/open/file/renameFile.action"
	idKey := "fileId"
	nameKey := "destFileName"
	if srcFile.IsDir() {
		url = "https://cloud.189.cn/api/open/file/renameFolder.action"
		idKey = "folderId"
		nameKey = "destFolderName"
	}
	form := map[string]string{
		idKey:   srcFile.Id,
		nameKey: dstName,
	}
	_, err = driver.Request(url, "POST", form, nil, account)
	return err
}

func (driver Cloud189) Copy(src string, dst string, account *model.Account) error {
	dstDir, dstName := filepath.Split(dst)
	srcFile, err := driver.File(src, account)
	if err != nil {
		return err
	}
	dstDirFile, err := driver.File(dstDir, account)
	if err != nil {
		return err
	}
	isFolder := 0
	if srcFile.IsDir() {
		isFolder = 1
	}
	taskInfos := []base.Json{
		{
			"fileId":   srcFile.Id,
			"fileName": dstName,
			"isFolder": isFolder,
		},
	}
	taskInfosBytes, err := json.Marshal(taskInfos)
	if err != nil {
		return err
	}
	form := map[string]string{
		"type":           "COPY",
		"targetFolderId": dstDirFile.Id,
		"taskInfos":      string(taskInfosBytes),
	}
	_, err = driver.Request("https://cloud.189.cn/api/open/batch/createBatchTask.action", "POST", form, nil, account)
	return err
}

func (driver Cloud189) Delete(path string, account *model.Account) error {
	path = utils.ParsePath(path)
	file, err := driver.File(path, account)
	if err != nil {
		return err
	}
	isFolder := 0
	if file.IsDir() {
		isFolder = 1
	}
	taskInfos := []base.Json{
		{
			"fileId":   file.Id,
			"fileName": file.Name,
			"isFolder": isFolder,
		},
	}
	taskInfosBytes, err := json.Marshal(taskInfos)
	if err != nil {
		return err
	}
	form := map[string]string{
		"type":           "DELETE",
		"targetFolderId": "",
		"taskInfos":      string(taskInfosBytes),
	}
	_, err = driver.Request("https://cloud.189.cn/api/open/batch/createBatchTask.action", "POST", form, nil, account)
	return err
}

// Upload Error: decrypt encryptionText failed
func (driver Cloud189) Upload(file *model.FileStream, account *model.Account) error {
	return base.ErrNotImplement
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
	res, err := driver.UploadRequest("/person/initMultiUpload", map[string]string{
		"parentFolderId": parentFile.Id,
		"fileName":       file.Name,
		"fileSize":       strconv.FormatInt(int64(file.Size), 10),
		"sliceSize":      strconv.FormatInt(int64(DEFAULT), 10),
		"lazyCheck":      "1",
	}, account)
	if err != nil {
		return err
	}
	uploadFileId := jsoniter.Get(res, "data.uploadFileId").ToString()
	var i int64
	var byteSize uint64
	md5s := make([]string, 0)
	md5Sum := md5.New()
	for i = 1; i <= count; i++ {
		byteSize = file.GetSize() - finish
		if DEFAULT < byteSize {
			byteSize = DEFAULT
		}
		log.Debugf("%d,%d", byteSize, finish)
		byteData := make([]byte, byteSize)
		n, err := io.ReadFull(file, byteData)
		log.Debug(err, n)
		if err != nil {
			return err
		}
		finish += uint64(n)
		md5Bytes := getMd5(byteData)
		md5Str := hex.EncodeToString(md5Bytes)
		md5Base64 := base64.StdEncoding.EncodeToString(md5Bytes)
		md5s = append(md5s, md5Str)
		md5Sum.Write(byteData)
		res, err = driver.UploadRequest("/person/getMultiUploadUrls", map[string]string{
			"partInfo":     fmt.Sprintf("%s-%s", strconv.FormatInt(i, 10), md5Base64),
			"uploadFileId": uploadFileId,
		}, account)
		if err != nil {
			return err
		}
		uploadData := jsoniter.Get(res, "uploadUrls.partNumber_"+strconv.FormatInt(i, 10))
		headers := strings.Split(uploadData.Get("requestHeader").ToString(), "&")
		req, err := http.NewRequest("PUT", uploadData.Get("requestURL").ToString(), bytes.NewBuffer(byteData))
		if err != nil {
			return err
		}
		for _, header := range headers {
			kv := strings.Split(header, "=")
			req.Header.Set(kv[0], strings.Join(kv[1:], "="))
		}
		res, err := base.HttpClient.Do(req)
		if err != nil {
			return err
		}
		log.Debugf("%+v", res)
	}
	id := md5Sum.Sum(nil)
	res, err = driver.UploadRequest("/person/commitMultiUploadFile", map[string]string{
		"uploadFileId": uploadFileId,
		"fileMd5":      hex.EncodeToString(id),
		"sliceMd5":     utils.GetMD5Encode(strings.Join(md5s, "\n")),
		"lazyCheck":    "1",
	}, account)
	if err == nil {
		_ = base.DeleteCache(file.ParentPath, account)
	}
	return err
}

var _ base.Driver = (*Cloud189)(nil)
