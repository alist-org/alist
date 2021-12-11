package _23

import (
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	url "net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Pan123 struct{}

func (driver Pan123) Config() base.DriverConfig {
	return base.DriverConfig{
		Name:      "123Pan",
		OnlyProxy: false,
	}
}

func (driver Pan123) Items() []base.Item {
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
			Required: false,
		},
		{
			Name:     "order_by",
			Label:    "order_by",
			Type:     base.TypeSelect,
			Values:   "name,fileId,updateAt,createAt",
			Required: true,
		},
		{
			Name:     "order_direction",
			Label:    "order_direction",
			Type:     base.TypeSelect,
			Values:   "asc,desc",
			Required: true,
		},
	}
}

func (driver Pan123) Save(account *model.Account, old *model.Account) error {
	if account.RootFolder == "" {
		account.RootFolder = "0"
	}
	err := driver.Login(account)
	return err
}

func (driver Pan123) File(path string, account *model.Account) (*model.File, error) {
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

func (driver Pan123) Files(path string, account *model.Account) ([]model.File, error) {
	path = utils.ParsePath(path)
	var rawFiles []Pan123File
	cache, err := base.GetCache(path, account)
	if err == nil {
		rawFiles, _ = cache.([]Pan123File)
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

func (driver Pan123) Link(path string, account *model.Account) (*base.Link, error) {
	file, err := driver.GetFile(utils.ParsePath(path), account)
	if err != nil {
		return nil, err
	}
	var resp Pan123DownResp
	_, err = pan123Client.R().SetResult(&resp).SetHeader("authorization", "Bearer "+account.AccessToken).
		SetBody(base.Json{
			"driveId":   0,
			"etag":      file.Etag,
			"fileId":    file.FileId,
			"fileName":  file.FileName,
			"s3keyFlag": file.S3KeyFlag,
			"size":      file.Size,
			"type":      file.Type,
		}).Post("https://www.123pan.com/api/file/download_info")
	if err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		if resp.Code == 401 {
			err := driver.Login(account)
			if err != nil {
				return nil, err
			}
			return driver.Link(path, account)
		}
		return nil, fmt.Errorf(resp.Message)
	}
	u, err := url.Parse(resp.Data.DownloadUrl)
	if err != nil {
		return nil, err
	}
	u_ := fmt.Sprintf("https://%s%s", u.Host, u.Path)
	res, err := base.NoRedirectClient.R().SetQueryParamsFromValues(u.Query()).Get(u_)
	if err != nil {
		return nil, err
	}
	log.Debug(res.String())
	link := base.Link{}
	if res.StatusCode() == 302 {
		link.Url = res.Header().Get("location")
	} else {
		link.Url = resp.Data.DownloadUrl
	}
	return &link, nil
}

func (driver Pan123) Path(path string, account *model.Account) (*model.File, []model.File, error) {
	path = utils.ParsePath(path)
	log.Debugf("pan123 path: %s", path)
	file, err := driver.File(path, account)
	if err != nil {
		return nil, nil, err
	}
	if !file.IsDir() {
		link, err := driver.Link(path, account)
		if err != nil {
			return nil, nil, err
		}
		file.Url = link.Url
		return file, nil, nil
	}
	files, err := driver.Files(path, account)
	if err != nil {
		return nil, nil, err
	}
	return nil, files, nil
}

func (driver Pan123) Proxy(c *gin.Context, account *model.Account) {
	c.Request.Header.Del("origin")
}

func (driver Pan123) Preview(path string, account *model.Account) (interface{}, error) {
	return nil, base.ErrNotSupport
}

func (driver Pan123) MakeDir(path string, account *model.Account) error {
	dir, name := filepath.Split(path)
	parentFile, err := driver.File(dir, account)
	if err != nil {
		return err
	}
	if !parentFile.IsDir() {
		return base.ErrNotFolder
	}
	parentFileId, _ := strconv.Atoi(parentFile.Id)
	data := base.Json{
		"driveId":      0,
		"etag":         "",
		"fileName":     name,
		"parentFileId": parentFileId,
		"size":         0,
		"type":         1,
	}
	_, err = driver.Post("https://www.123pan.com/api/file/upload_request", data, account)
	if err == nil {
		_ = base.DeleteCache(dir, account)
	}
	return err
}

func (driver Pan123) Move(src string, dst string, account *model.Account) error {
	srcDir, _ := filepath.Split(src)
	dstDir, dstName := filepath.Split(dst)
	srcFile, err := driver.File(src, account)
	if err != nil {
		return err
	}
	fileId, _ := strconv.Atoi(srcFile.Id)
	// rename
	if srcDir == dstDir {
		data := base.Json{
			"driveId":  0,
			"fileId":   fileId,
			"fileName": dstName,
		}
		_, err = driver.Post("https://www.123pan.com/api/file/rename", data, account)
	} else {
		// move
		dstDirFile, err := driver.File(dstDir, account)
		if err != nil {
			return err
		}
		parentFileId, _ := strconv.Atoi(dstDirFile.Id)
		data := base.Json{
			"fileId":       fileId,
			"parentFileId": parentFileId,
		}
		_, err = driver.Post("https://www.123pan.com/api/file/mod_pid", data, account)
	}
	if err != nil {
		_ = base.DeleteCache(srcDir, account)
		_ = base.DeleteCache(dstDir, account)
	}
	return err
}

func (driver Pan123) Copy(src string, dst string, account *model.Account) error {
	return base.ErrNotSupport
}

func (driver Pan123) Delete(path string, account *model.Account) error {
	file, err := driver.GetFile(path, account)
	if err != nil {
		return err
	}
	data := base.Json{
		"driveId":           0,
		"operation":         true,
		"fileTrashInfoList": file,
	}
	_, err = driver.Post("https://www.123pan.com/api/file/trash", data, account)
	if err == nil {
		_ = base.DeleteCache(utils.Dir(path), account)
	}
	return err
}

type UploadResp struct {
	XMLName  xml.Name `xml:"InitiateMultipartUploadResult"`
	Bucket   string   `xml:"Bucket"`
	Key      string   `xml:"Key"`
	UploadId string   `xml:"UploadId"`
}

// TODO unfinished
func (driver Pan123) Upload(file *model.FileStream, account *model.Account) error {
	parentFile, err := driver.File(file.ParentPath, account)
	if err != nil {
		return err
	}
	if !parentFile.IsDir() {
		return base.ErrNotFolder
	}
	parentFileId, _ := strconv.Atoi(parentFile.Id)
	data := base.Json{
		"driveId":      0,
		"duplicate":    true,
		"etag":         RandStr(32), //maybe file's md5
		"fileName":     file.GetFileName(),
		"parentFileId": parentFileId,
		"size":         file.GetSize(),
		"type":         0,
	}
	res, err := driver.Post("https://www.123pan.com/api/file/upload_request", data, account)
	if err != nil {
		return err
	}
	baseUrl := fmt.Sprintf("https://file.123pan.com/%s/%s", jsoniter.Get(res, "data.Bucket").ToString(), jsoniter.Get(res, "data.Key").ToString())
	var resp UploadResp
	kSecret := jsoniter.Get(res, "data.SecretAccessKey").ToString()
	nowTimeStr := time.Now().String()
	Date := strings.ReplaceAll(strings.Split(nowTimeStr, "T")[0],"-","")

	StringToSign := fmt.Sprintf("%s\n%s\n%s\n%s",
		"AWS4-HMAC-SHA256",
		nowTimeStr,
		fmt.Sprintf("%s/us-east-1/s3/aws4_request", Date),
		)

	kDate := HMAC("AWS4"+kSecret, Date)
	kRegion := HMAC(kDate, "us-east-1")
	kService := HMAC(kRegion, "s3")
	kSigning := HMAC(kService, "aws4_request")
	_, err = pan123Client.R().SetResult(&resp).SetHeaders(map[string]string{
		"Authorization": fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s/us-east-1/s3/aws4_request, SignedHeaders=host;x-amz-content-sha256;x-amz-date;x-amz-security-token;x-amz-user-agent, Signature=%s",
			jsoniter.Get(res, "data.AccessKeyId"),
			Date,
			hex.EncodeToString([]byte(HMAC(StringToSign, kSigning)))),
		"X-Amz-Content-Sha256": "UNSIGNED-PAYLOAD",
		"X-Amz-Date":           nowTimeStr,
		"x-amz-security-token": jsoniter.Get(res, "data.SessionToken").ToString(),
	}).Post(fmt.Sprintf("%s?uploads", baseUrl))
	if err != nil {
		return err
	}
	return base.ErrNotImplement
}

var _ base.Driver = (*Pan123)(nil)
