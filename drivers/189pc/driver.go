package _189

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

func init() {
	base.RegisterDriver(new(Cloud189))
}

type Cloud189 struct {
}

func (driver Cloud189) Config() base.DriverConfig {
	return base.DriverConfig{
		Name: "189CloudPC",
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
			Name:     "internal_type",
			Label:    "189cloud type",
			Type:     base.TypeSelect,
			Required: true,
			Values:   "Personal,Family",
		},
		{
			Name:  "site_id",
			Label: "family id",
			Type:  base.TypeString,
		},
		{
			Name:     "order_by",
			Label:    "order_by",
			Type:     base.TypeSelect,
			Values:   "filename,filesize,lastOpTime",
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
	if account == nil {
		return nil
	}

	if !isFamily(account) && account.RootFolder == "" {
		account.RootFolder = "-11"
	}

	state := GetState(account)
	if !state.IsLogin() {
		return state.Login(account)
	}
	account.Status = "work"
	model.SaveAccount(account)
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
	cache, err := base.GetCache(path, account)
	if err == nil {
		files, _ := cache.([]model.File)
		return files, nil
	}

	file, err := driver.File(path, account)
	if err != nil {
		return nil, err
	}

	fullUrl := API_URL
	if isFamily(account) {
		fullUrl += "/family/file"
	}
	fullUrl += "/listFiles.action"

	files := make([]model.File, 0)
	client := GetState(account)
	for pageNum := 1; ; pageNum++ {
		var resp Cloud189FilesResp
		queryparam := map[string]string{
			"folderId":   file.Id,
			"fileType":   "0",
			"mediaAttr":  "0",
			"iconOption": "5",
			"pageNum":    fmt.Sprint(pageNum),
			"pageSize":   "130",
		}
		_, err = client.Request("GET", fullUrl, nil, func(r *resty.Request) {
			r.SetQueryParams(clientSuffix()).SetQueryParams(queryparam)
			if isFamily(account) {
				r.SetQueryParams(map[string]string{
					"familyId":   account.SiteId,
					"orderBy":    toFamilyOrderBy(account.OrderBy),
					"descending": account.OrderDirection,
				})
			} else {
				r.SetQueryParams(map[string]string{
					"recursive":  "0",
					"orderBy":    account.OrderBy,
					"descending": account.OrderDirection,
				})
			}
			r.SetResult(&resp)
		}, account)
		if err != nil {
			return nil, err
		}
		// 获取完毕跳出
		if resp.FileListAO.Count == 0 {
			break
		}

		mustTime := func(str string) *time.Time {
			time, _ := http.ParseTime(str)
			return &time
		}
		for _, folder := range resp.FileListAO.FolderList {
			files = append(files, model.File{
				Id:        fmt.Sprint(folder.ID),
				Name:      folder.Name,
				Size:      0,
				Type:      conf.FOLDER,
				Driver:    driver.Config().Name,
				UpdatedAt: mustTime(folder.CreateDate),
			})
		}
		for _, file := range resp.FileListAO.FileList {
			files = append(files, model.File{
				Id:        fmt.Sprint(file.ID),
				Name:      file.Name,
				Size:      file.Size,
				Type:      utils.GetFileType(filepath.Ext(file.Name)),
				Driver:    driver.Config().Name,
				UpdatedAt: mustTime(file.CreateDate),
				Thumbnail: file.Icon.SmallUrl,
			})
		}
	}
	if len(files) > 0 {
		_ = base.SetCache(path, files, account)
	}
	return files, nil
}

func (driver Cloud189) Path(path string, account *model.Account) (*model.File, []model.File, error) {
	path = utils.ParsePath(path)
	log.Debugf("189PC path: %s", path)
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

func (driver Cloud189) Link(args base.Args, account *model.Account) (*base.Link, error) {
	file, err := driver.File(utils.ParsePath(args.Path), account)
	if err != nil {
		return nil, err
	}
	if file.Type == conf.FOLDER {
		return nil, base.ErrNotFile
	}

	fullUrl := API_URL
	if isFamily(account) {
		fullUrl += "/family/file"
	}
	fullUrl += "/getFileDownloadUrl.action"

	var downloadUrl struct {
		URL string `json:"fileDownloadUrl"`
	}
	_, err = GetState(account).Request("GET", fullUrl, nil, func(r *resty.Request) {
		r.SetQueryParams(clientSuffix()).SetQueryParam("fileId", file.Id)
		if isFamily(account) {
			r.SetQueryParams(map[string]string{
				"familyId": account.SiteId,
			})
		} else {
			r.SetQueryParams(map[string]string{
				"dt":   "3",
				"flag": "1",
			})
		}
		r.SetResult(&downloadUrl)
	}, account)
	if err != nil {
		return nil, err
	}
	return &base.Link{
		Headers: []base.Header{
			{Name: "User-Agent", Value: base.UserAgent},
		},
		Url: strings.ReplaceAll(downloadUrl.URL, "&amp;", "&"),
	}, nil
}

func (driver Cloud189) Preview(path string, account *model.Account) (interface{}, error) {
	return nil, base.ErrNotSupport
}

func (driver Cloud189) MakeDir(path string, account *model.Account) error {
	dir, name := filepath.Split(path)
	parentFile, err := driver.File(dir, account)
	if err != nil {
		return err
	}
	if !parentFile.IsDir() {
		return base.ErrNotFolder
	}

	fullUrl := API_URL
	if isFamily(account) {
		fullUrl += "/family/file"
	}
	fullUrl += "/createFolder.action"

	_, err = GetState(account).Request("POST", fullUrl, nil, func(r *resty.Request) {
		r.SetQueryParams(clientSuffix()).SetQueryParams(map[string]string{
			"folderName":   name,
			"relativePath": "",
		})
		if isFamily(account) {
			r.SetQueryParams(map[string]string{
				"familyId": account.SiteId,
				"parentId": parentFile.Id,
			})
		} else {
			r.SetQueryParams(map[string]string{
				"parentFolderId": parentFile.Id,
			})
		}
	}, account)
	return err
}

func (driver Cloud189) Move(src string, dst string, account *model.Account) error {
	srcFile, err := driver.File(src, account)
	if err != nil {
		return err
	}

	dstDirFile, err := driver.File(filepath.Dir(dst), account)
	if err != nil {
		return err
	}

	_, err = GetState(account).Request("POST", API_URL+"/batch/createBatchTask.action", nil, func(r *resty.Request) {
		r.SetFormData(clientSuffix()).SetFormData(map[string]string{
			"type": "MOVE",
			"taskInfos": string(MustToBytes(json.Marshal(
				[]*BatchTaskInfo{
					{
						FileId:   srcFile.Id,
						FileName: srcFile.Name,
						IsFolder: BoolToNumber(srcFile.IsDir()),
					},
				}))),
			"targetFolderId": dstDirFile.Id,
		})
		if isFamily(account) {
			r.SetFormData(map[string]string{
				"familyId": account.SiteId,
			})
		}
	}, account)
	return err
}

/*
func (driver Cloud189) Move(src string, dst string, account *model.Account) error {
	srcFile, err := driver.File(src, account)
	if err != nil {
		return err
	}

	dstDirFile, err := driver.File(filepath.Dir(dst), account)
	if err != nil {
		return err
	}

	var queryParam map[string]string
	fullUrl := API_URL
	method := "POST"
	if isFamily(account) {
		fullUrl += "/family/file"
		method = "GET"
	}
	if srcFile.IsDir() {
		fullUrl += "/moveFolder.action"
		queryParam = map[string]string{
			"folderId":       srcFile.Id,
			"destFolderName": srcFile.Name,
		}
	} else {
		fullUrl += "/moveFile.action"
		queryParam = map[string]string{
			"fileId":       srcFile.Id,
			"destFileName": srcFile.Name,
		}
	}

	_, err = GetState(account).Request(method, fullUrl, nil, func(r *resty.Request) {
		r.SetQueryParams(queryParam).SetQueryParams(clientSuffix())
		if isFamily(account) {
			r.SetQueryParams(map[string]string{
				"familyId":     account.SiteId,
				"destParentId": dstDirFile.Id,
			})
		} else {
			r.SetQueryParam("destParentFolderId", dstDirFile.Id)
		}
	}, account)
	return err
}*/

func (driver Cloud189) Rename(src string, dst string, account *model.Account) error {
	srcFile, err := driver.File(src, account)
	if err != nil {
		return err
	}

	var queryParam map[string]string
	fullUrl := API_URL
	method := "POST"
	if isFamily(account) {
		fullUrl += "/family/file"
		method = "GET"
	}
	if srcFile.IsDir() {
		fullUrl += "/renameFolder.action"
		queryParam = map[string]string{
			"folderId":       srcFile.Id,
			"destFolderName": filepath.Base(dst),
		}
	} else {
		fullUrl += "/renameFile.action"
		queryParam = map[string]string{
			"fileId":       srcFile.Id,
			"destFileName": filepath.Base(dst),
		}
	}

	_, err = GetState(account).Request(method, fullUrl, nil, func(r *resty.Request) {
		r.SetQueryParams(queryParam).SetQueryParams(clientSuffix())
		if isFamily(account) {
			r.SetQueryParam("familyId", account.SiteId)
		}
	}, account)
	return err
}

func (driver Cloud189) Copy(src string, dst string, account *model.Account) error {
	srcFile, err := driver.File(src, account)
	if err != nil {
		return err
	}

	dstDirFile, err := driver.File(filepath.Dir(dst), account)
	if err != nil {
		return err
	}

	_, err = GetState(account).Request("POST", API_URL+"/batch/createBatchTask.action", nil, func(r *resty.Request) {
		r.SetFormData(clientSuffix()).SetFormData(map[string]string{
			"type": "COPY",
			"taskInfos": string(MustToBytes(json.Marshal(
				[]*BatchTaskInfo{
					{
						FileId:   srcFile.Id,
						FileName: srcFile.Name,
						IsFolder: BoolToNumber(srcFile.IsDir()),
					},
				}))),
			"targetFolderId": dstDirFile.Id,
			"targetFileName": filepath.Base(dst),
		})
		if isFamily(account) {
			r.SetFormData(map[string]string{
				"familyId": account.SiteId,
			})
		}
	}, account)
	return err
}

func (driver Cloud189) Delete(path string, account *model.Account) error {
	path = utils.ParsePath(path)
	srcFile, err := driver.File(path, account)
	if err != nil {
		return err
	}

	_, err = GetState(account).Request("POST", API_URL+"/batch/createBatchTask.action", nil, func(r *resty.Request) {
		r.SetFormData(clientSuffix()).SetFormData(map[string]string{
			"type": "DELETE",
			"taskInfos": string(MustToBytes(json.Marshal(
				[]*BatchTaskInfo{
					{
						FileId:   srcFile.Id,
						FileName: srcFile.Name,
						IsFolder: BoolToNumber(srcFile.IsDir()),
					},
				}))),
		})

		if isFamily(account) {
			r.SetFormData(map[string]string{
				"familyId": account.SiteId,
			})
		}
	}, account)
	return err
}

func (driver Cloud189) Upload(file *model.FileStream, account *model.Account) error {
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

	if isFamily(account) {
		return driver.uploadFamily(file, parentFile, account)
	}
	return driver.uploadPerson(file, parentFile, account)
}

func (driver Cloud189) uploadFamily(file *model.FileStream, parentFile *model.File, account *model.Account) error {
	tempFile, err := ioutil.TempFile(conf.Conf.TempDir, "file-*")
	if err != nil {
		return err
	}

	defer tempFile.Close()
	defer os.Remove(tempFile.Name())

	fileMd5 := md5.New()
	if _, err = io.Copy(io.MultiWriter(fileMd5, tempFile), file); err != nil {
		return err
	}

	client := GetState(account)
	var createUpload CreateUploadFileResult
	_, err = client.Request("GET", API_URL+"/family/file/createFamilyFile.action", nil, func(r *resty.Request) {
		r.SetQueryParams(map[string]string{
			"fileMd5":      hex.EncodeToString(fileMd5.Sum(nil)),
			"fileName":     file.Name,
			"familyId":     account.SiteId,
			"parentId":     parentFile.Id,
			"resumePolicy": "1",
			"fileSize":     fmt.Sprint(file.Size),
		})
		r.SetQueryParams(clientSuffix())
		r.SetResult(&createUpload)
	}, account)
	if err != nil {
		return err
	}

	if createUpload.FileDataExists != 1 {
		if err = driver.uploadFileData(file, tempFile, createUpload, account); err != nil {
			return err
		}
	}

	_, err = client.Request("GET", createUpload.FileCommitUrl, nil, func(r *resty.Request) {
		r.SetQueryParams(clientSuffix())
		r.SetHeaders(map[string]string{
			"FamilyId":     account.SiteId,
			"uploadFileId": fmt.Sprint(createUpload.UploadFileId),
			"ResumePolicy": "1",
		})
	}, account)
	return err
}

func (driver Cloud189) uploadPerson(file *model.FileStream, parentFile *model.File, account *model.Account) error {
	tempFile, err := ioutil.TempFile(conf.Conf.TempDir, "file-*")
	if err != nil {
		return err
	}

	defer tempFile.Close()
	defer os.Remove(tempFile.Name())

	fileMd5 := md5.New()
	if _, err = io.Copy(io.MultiWriter(fileMd5, tempFile), file); err != nil {
		return err
	}

	client := GetState(account)
	var createUpload CreateUploadFileResult
	_, err = client.Request("POST", API_URL+"/createUploadFile.action", nil, func(r *resty.Request) {
		r.SetQueryParams(clientSuffix())
		r.SetFormData(clientSuffix()).SetFormData(map[string]string{
			"parentFolderId": parentFile.Id,
			"baseFileId":     "",
			"fileName":       file.Name,
			"size":           fmt.Sprint(file.Size),
			"md5":            hex.EncodeToString(fileMd5.Sum(nil)),
			//  "lastWrite":      param.LastWrite,
			//	"localPath":      strings.ReplaceAll(file.ParentPath, "\\", "/"),
			"opertype":     "1",
			"flag":         "1",
			"resumePolicy": "1",
			"isLog":        "0",
			"fileExt":      "",
		})
		r.SetResult(&createUpload)
	}, account)
	if err != nil {
		return err
	}

	if createUpload.FileDataExists != 1 {
		if err = driver.uploadFileData(file, tempFile, createUpload, account); err != nil {
			return err
		}
	}

	_, err = client.Request("POST", createUpload.FileCommitUrl, nil, func(r *resty.Request) {
		r.SetQueryParams(clientSuffix())
		r.SetFormData(map[string]string{
			"uploadFileId": fmt.Sprint(createUpload.UploadFileId),
			"opertype":     "1", //5 覆盖
			"ResumePolicy": "1",
			"isLog":        "0",
		})
	}, account)
	return err
}

func (driver Cloud189) uploadFileData(file *model.FileStream, tempFile *os.File, createUpload CreateUploadFileResult, account *model.Account) error {
	var uploadFileState *UploadFileStatusResult
	var err error
	for i := 0; i < 10; i++ {
		if uploadFileState, err = driver.getUploadFileState(createUpload.UploadFileId, account); err != nil {
			return err
		}

		if uploadFileState.FileDataExists == 1 || uploadFileState.DataSize == int64(file.Size) {
			return nil
		}

		if _, err = tempFile.Seek(uploadFileState.DataSize, io.SeekStart); err != nil {
			return err
		}

		_, err = GetState(account).Request("PUT", uploadFileState.FileUploadUrl, nil, func(r *resty.Request) {
			r.SetQueryParams(clientSuffix())
			r.SetHeaders(map[string]string{
				"ResumePolicy":           "1",
				"Edrive-UploadFileId":    fmt.Sprint(createUpload.UploadFileId),
				"Edrive-UploadFileRange": fmt.Sprintf("bytes=%d-%d", uploadFileState.DataSize, file.Size),
				"Expect":                 "100-continue",
			})
			if isFamily(account) {
				r.SetHeader("FamilyId", account.SiteId)
			}
			r.SetBody(tempFile)
		}, account)
		if err == nil {
			break
		}
	}
	return err
}

func (driver Cloud189) getUploadFileState(uploadFileId int64, account *model.Account) (*UploadFileStatusResult, error) {
	fullUrl := API_URL
	if isFamily(account) {
		fullUrl += "/family/file/getFamilyFileStatus.action"
	} else {
		fullUrl += "/getUploadFileStatus.action"
	}
	var uploadFileState UploadFileStatusResult
	_, err := GetState(account).Request("GET", fullUrl, nil, func(r *resty.Request) {
		r.SetQueryParams(clientSuffix())
		r.SetQueryParams(map[string]string{
			"uploadFileId": fmt.Sprint(uploadFileId),
			"resumePolicy": "1",
		})
		if isFamily(account) {
			r.SetQueryParam("familyId", account.SiteId)
		}
		r.SetResult(&uploadFileState)
	}, account)
	if err != nil {
		return nil, err
	}
	return &uploadFileState, nil
}

/*
暂时未解决
func (driver Cloud189) Upload(file *model.FileStream, account *model.Account) error {
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

	fullUrl := UPLOAD_URL
	if isFamily(account) {
		fullUrl += "/family"
	} else {
		fullUrl += "/person"
	}

	tempFile, err := ioutil.TempFile(conf.Conf.TempDir, "file-*")
	if err != nil {
		return err
	}

	defer tempFile.Close()
	defer os.Remove(tempFile.Name())

	// 初始化上传
	const DEFAULT int64 = 10485760
	count := int64(math.Ceil(float64(file.Size) / float64(DEFAULT)))
	fileMd5 := md5.New()
	silceMd5 := md5.New()
	silceMd5Hexs := make([]string, 0, count)
	silceMd5Base64s := make([]string, 0, count)
	for i := int64(1); i <= count; i++ {
		if _, err := io.CopyN(io.MultiWriter(fileMd5, silceMd5, tempFile), file, DEFAULT); err != io.EOF {
			return err
		}
		md5Byte := silceMd5.Sum(nil)
		silceMd5Hexs = append(silceMd5Hexs, strings.ToUpper(hex.EncodeToString(md5Byte)))
		silceMd5Base64s = append(silceMd5Base64s, fmt.Sprint(i, "-", base64.StdEncoding.EncodeToString(md5Byte)))
	}
	fileMd5Hex := strings.ToUpper(hex.EncodeToString(fileMd5.Sum(nil)))
	sliceMd5Hex := fileMd5Hex
	if int64(file.Size) > DEFAULT {
		sliceMd5Hex = strings.ToUpper(utils.GetMD5Encode(strings.Join(silceMd5Hexs, "\n")))
	}

	qID := uuid.NewString()
	client := GetState(account)
	param := MapToUrlValues(map[string]interface{}{
		"parentFolderId": parentFile.Id,
		"fileName":       url.QueryEscape(file.Name),
		"fileMd5":        fileMd5Hex,
		"fileSize":       fmt.Sprint(file.Size),
		"sliceMd5":       sliceMd5Hex,
		"sliceSize":      fmt.Sprint(DEFAULT),
	})
	if isFamily(account) {
		param.Set("familyId", account.SiteId)
	}

	var uploadInfo InitMultiUploadResp
	_, err = client.Request("GET", fullUrl+"/initMultiUpload", param, func(r *resty.Request) {
		r.SetQueryParams(clientSuffix())
		r.SetHeader("X-Request-ID", qID)
		r.SetResult(&uploadInfo)
	}, account)
	if err != nil {
		return err
	}

	if uploadInfo.Data.FileDataExists != 1 {
		param = MapToUrlValues(map[string]interface{}{
			"uploadFileId": uploadInfo.Data.UploadFileID,
			"partInfo":     strings.Join(silceMd5Base64s, ","),
		})
		if isFamily(account) {
			param.Set("familyId", account.SiteId)
		}
		var uploadUrls UploadUrlsResp
		_, err := client.Request("GET", fullUrl+"/getMultiUploadUrls", param, func(r *resty.Request) {
			r.SetQueryParams(clientSuffix())
			r.SetHeader("X-Request-ID", qID).SetHeader("content-type", "application/x-www-form-urlencoded")
			r.SetResult(&uploadUrls)

		}, account)
		if err != nil {
			return err
		}
		var i int64
		for _, uploadurl := range uploadUrls.UploadUrls {
			req := resty.New().SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}).SetProxy("http://192.168.0.30:8888").R()
			for _, header := range strings.Split(decodeURIComponent(uploadurl.RequestHeader), "&") {
				i := strings.Index(header, "=")
				req.SetHeader(header[0:i], header[i+1:])
			}
			_, err := req.SetBody(io.NewSectionReader(tempFile, i*DEFAULT, DEFAULT)).Put(uploadurl.RequestURL)
			if err != nil {
				return err
			}
		}
	}

	param = MapToUrlValues(map[string]interface{}{
		"uploadFileId": uploadInfo.Data.UploadFileID,
		"isLog":        "0",
		"opertype":     "1",
	})
	if isFamily(account) {
		param.Set("familyId", account.SiteId)
	}
	_, err = client.Request("GET", fullUrl+"/commitMultiUploadFile", param, func(r *resty.Request) {
		r.SetHeader("X-Request-ID", qID)
		r.SetQueryParams(clientSuffix())
	}, account)
	return err
}
*/
var _ base.Driver = (*Cloud189)(nil)
