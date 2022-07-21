package _23

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	log "github.com/sirupsen/logrus"
)

type Pan123 struct{}

func (driver Pan123) Config() base.DriverConfig {
	return base.DriverConfig{
		Name: "123Pan",
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
			Default:  "name",
		},
		{
			Name:     "order_direction",
			Label:    "order_direction",
			Type:     base.TypeSelect,
			Values:   "asc,desc",
			Required: true,
			Default:  "asc",
		},
		{
			Name:        "bool_1",
			Label:       "stream upload",
			Type:        base.TypeBool,
			Description: "io stream upload (test)",
		},
	}
}

func (driver Pan123) Save(account *model.Account, old *model.Account) error {
	if account == nil {
		return nil
	}
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
	var rawFiles []File
	cache, err := base.GetCache(path, account)
	if err == nil {
		rawFiles, _ = cache.([]File)
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
	files := make([]model.File, 0, len(rawFiles))
	for _, file := range rawFiles {
		files = append(files, *driver.FormatFile(&file))
	}
	return files, nil
}

func (driver Pan123) Link(args base.Args, account *model.Account) (*base.Link, error) {
	log.Debugf("%+v", args)
	file, err := driver.GetFile(utils.ParsePath(args.Path), account)
	if err != nil {
		return nil, err
	}
	var resp Pan123DownResp
	var headers map[string]string
	if !utils.IsLocalIPAddr(args.IP) {
		headers = map[string]string{
			//"X-Real-IP":       "1.1.1.1",
			"X-Forwarded-For": args.IP,
		}
	}
	data := base.Json{
		"driveId":   0,
		"etag":      file.Etag,
		"fileId":    file.FileId,
		"fileName":  file.FileName,
		"s3keyFlag": file.S3KeyFlag,
		"size":      file.Size,
		"type":      file.Type,
	}
	_, err = driver.Request("https://www.123pan.com/api/file/download_info",
		base.Post, headers, nil, &data, &resp, false, account)
	//_, err = pan123Client.R().SetResult(&resp).SetHeader("authorization", "Bearer "+account.AccessToken).
	//	SetBody().Post("https://www.123pan.com/api/file/download_info")
	if err != nil {
		return nil, err
	}
	u, err := url.Parse(resp.Data.DownloadUrl)
	if err != nil {
		return nil, err
	}
	u_ := fmt.Sprintf("https://%s%s", u.Host, u.Path)
	res, err := base.NoRedirectClient.R().SetQueryParamsFromValues(u.Query()).Head(u_)
	if err != nil {
		return nil, err
	}
	log.Debug(res.String())
	link := base.Link{
		Url: resp.Data.DownloadUrl,
	}
	log.Debugln("res code: ", res.StatusCode())
	if res.StatusCode() == 302 {
		link.Url = res.Header().Get("location")
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
		return file, nil, nil
	}
	files, err := driver.Files(path, account)
	if err != nil {
		return nil, nil, err
	}
	return nil, files, nil
}

//func (driver Pan123) Proxy(r *http.Request, account *model.Account) {
//	r.Header.Del("origin")
//}

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
	_, err = driver.Request("https://www.123pan.com/api/file/upload_request",
		base.Post, nil, nil, &data, nil, false, account)
	return err
}

func (driver Pan123) Move(src string, dst string, account *model.Account) error {
	dstDir, _ := filepath.Split(dst)
	srcFile, err := driver.File(src, account)
	if err != nil {
		return err
	}
	fileId, _ := strconv.Atoi(srcFile.Id)

	dstDirFile, err := driver.File(dstDir, account)
	if err != nil {
		return err
	}
	parentFileId, _ := strconv.Atoi(dstDirFile.Id)
	data := base.Json{
		"fileIdList":   []base.Json{{"FileId": fileId}},
		"parentFileId": parentFileId,
	}
	_, err = driver.Request("https://www.123pan.com/api/file/mod_pid",
		base.Post, nil, nil, &data, nil, false, account)
	return err
}

func (driver Pan123) Rename(src string, dst string, account *model.Account) error {
	_, dstName := filepath.Split(dst)
	srcFile, err := driver.File(src, account)
	if err != nil {
		return err
	}
	fileId, _ := strconv.Atoi(srcFile.Id)

	data := base.Json{
		"driveId":  0,
		"fileId":   fileId,
		"fileName": dstName,
	}
	_, err = driver.Request("https://www.123pan.com/api/file/rename",
		base.Post, nil, nil, &data, nil, false, account)
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
	log.Debugln("delete 123 file: ", file)
	data := base.Json{
		"driveId":           0,
		"operation":         true,
		"fileTrashInfoList": []File{*file},
	}
	_, err = driver.Request("https://www.123pan.com/b/api/file/trash",
		base.Post, nil, nil, &data, nil, false, account)
	return err
}

func (driver Pan123) Upload(file *model.FileStream, account *model.Account) error {
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
	var uploadFile io.Reader
	h := md5.New()
	if account.Bool1 && file.GetSize() > uint64(DEFAULT) {
		// 只计算前10MIB
		buf := bytes.NewBuffer(make([]byte, 0, DEFAULT))
		if n, err := io.CopyN(io.MultiWriter(buf, h), file, DEFAULT); err != io.EOF && n == 0 {
			return err
		}
		// 增加额外参数防止MD5碰撞
		h.Write([]byte(file.Name))
		num := make([]byte, 8)
		binary.BigEndian.PutUint64(num, file.Size)
		h.Write(num)
		// 拼装
		uploadFile = io.MultiReader(buf, file)
	} else {
		// 计算完整文件MD5
		tempFile, err := ioutil.TempFile(conf.Conf.TempDir, "file-*")
		if err != nil {
			return err
		}
		defer func() {
			_ = tempFile.Close()
			_ = os.Remove(tempFile.Name())
		}()

		if _, err = io.Copy(io.MultiWriter(tempFile, h), file); err != nil {
			return err
		}

		_, err = tempFile.Seek(0, io.SeekStart)
		if err != nil {
			return err
		}
		uploadFile = tempFile
	}
	etag := hex.EncodeToString(h.Sum(nil))
	data := base.Json{
		"driveId":      0,
		"duplicate":    2, // 2->覆盖 1->重命名 0->默认
		"etag":         etag,
		"fileName":     file.GetFileName(),
		"parentFileId": parentFile.Id,
		"size":         file.GetSize(),
		"type":         0,
	}
	var resp UploadResp
	_, err = driver.Request("https://www.123pan.com/api/file/upload_request",
		base.Post, map[string]string{"app-version": "1.1"}, nil, &data, &resp, false, account)
	//res, err := driver.Post("https://www.123pan.com/api/file/upload_request", data, account)
	if err != nil {
		return err
	}
	if resp.Data.Key == "" {
		return nil
	}
	cfg := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(resp.Data.AccessKeyId, resp.Data.SecretAccessKey, resp.Data.SessionToken),
		Region:           aws.String("123pan"),
		Endpoint:         aws.String("file.123pan.com"),
		S3ForcePathStyle: aws.Bool(true),
	}
	s, err := session.NewSession(cfg)
	if err != nil {
		return err
	}
	uploader := s3manager.NewUploader(s)
	input := &s3manager.UploadInput{
		Bucket: &resp.Data.Bucket,
		Key:    &resp.Data.Key,
		Body:   uploadFile,
	}
	_, err = uploader.Upload(input)
	if err != nil {
		return err
	}
	_, err = driver.Request("https://www.123pan.com/api/file/upload_complete", base.Post, nil, nil, &base.Json{
		"fileId": resp.Data.FileId,
	}, nil, false, account)
	return err
}

//type UploadResp struct {
//	XMLName  xml.Name `xml:"InitiateMultipartUploadResult"`
//	Bucket   string   `xml:"Bucket"`
//	Key      string   `xml:"Key"`
//	UploadId string   `xml:"UploadId"`
//}

// TODO unfinished
//func (driver Pan123) Upload(file *model.FileStream, account *model.Account) error {
//	return base.ErrNotImplement
//	parentFile, err := driver.File(file.ParentPath, account)
//	if err != nil {
//		return err
//	}
//	if !parentFile.IsDir() {
//		return base.ErrNotFolder
//	}
//	parentFileId, _ := strconv.Atoi(parentFile.Id)
//	data := base.Json{
//		"driveId":      0,
//		"duplicate":    true,
//		"etag":         RandStr(32), //maybe file's md5
//		"fileName":     file.GetFileName(),
//		"parentFileId": parentFileId,
//		"size":         file.GetSize(),
//		"type":         0,
//	}
//	res, err := driver.Request("https://www.123pan.com/api/file/upload_request",
//		base.Post, nil, nil, &data, nil, false, account)
//	//res, err := driver.Post("https://www.123pan.com/api/file/upload_request", data, account)
//	if err != nil {
//		return err
//	}
//	baseUrl := fmt.Sprintf("https://file.123pan.com/%s/%s", jsoniter.Get(res, "data.Bucket").ToString(), jsoniter.Get(res, "data.Key").ToString())
//	var resp UploadResp
//	kSecret := jsoniter.Get(res, "data.SecretAccessKey").ToString()
//	nowTimeStr := time.Now().String()
//	Date := strings.ReplaceAll(strings.Split(nowTimeStr, "T")[0], "-", "")
//
//	StringToSign := fmt.Sprintf("%s\n%s\n%s\n%s",
//		"AWS4-HMAC-SHA256",
//		nowTimeStr,
//		fmt.Sprintf("%s/us-east-1/s3/aws4_request", Date),
//	)
//
//	kDate := HMAC("AWS4"+kSecret, Date)
//	kRegion := HMAC(kDate, "us-east-1")
//	kService := HMAC(kRegion, "s3")
//	kSigning := HMAC(kService, "aws4_request")
//	_, err = base.RestyClient.R().SetResult(&resp).SetHeaders(map[string]string{
//		"Authorization": fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s/us-east-1/s3/aws4_request, SignedHeaders=host;x-amz-content-sha256;x-amz-date;x-amz-security-token;x-amz-user-agent, Signature=%s",
//			jsoniter.Get(res, "data.AccessKeyId"),
//			Date,
//			hex.EncodeToString([]byte(HMAC(StringToSign, kSigning)))),
//		"X-Amz-Content-Sha256": "UNSIGNED-PAYLOAD",
//		"X-Amz-Date":           nowTimeStr,
//		"x-amz-security-token": jsoniter.Get(res, "data.SessionToken").ToString(),
//	}).Post(fmt.Sprintf("%s?uploads", baseUrl))
//	if err != nil {
//		return err
//	}
//	return base.ErrNotImplement
//}

var _ base.Driver = (*Pan123)(nil)
