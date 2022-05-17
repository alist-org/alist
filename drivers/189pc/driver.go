package _189

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

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
			Name:  "root_folder",
			Label: "root folder file_id",
			Type:  base.TypeString,
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
		{
			Name:  "bool_1",
			Label: "fast upload",
			Type:  base.TypeBool,
		},
	}
}

func (driver Cloud189) Save(account *model.Account, old *model.Account) error {
	if account == nil {
		return nil
	}

	if !isFamily(account) && account.RootFolder == "" {
		account.RootFolder = "-11"
		account.SiteId = ""
	}
	if isFamily(account) && account.RootFolder == "-11" {
		account.RootFolder = ""
	}

	state := GetState(account)
	if !state.IsLogin(account) {
		if err := state.Login(account); err != nil {
			return err
		}
	}

	if isFamily(account) {
		list, err := driver.getFamilyInfoList(account)
		if err != nil {
			return err
		}
		for _, l := range list {
			if account.SiteId == "" {
				account.SiteId = fmt.Sprint(l.FamilyID)
			}
			log.Infof("天翼家庭云 用户名：%s FamilyID %d\n", l.RemarkName, l.FamilyID)
		}
	}

	account.Status = "work"
	model.SaveAccount(account)
	return nil
}

func (driver Cloud189) getFamilyInfoList(account *model.Account) ([]FamilyInfoResp, error) {
	var resp FamilyInfoListResp
	_, err := GetState(account).Request(http.MethodGet, API_URL+"/family/manage/getFamilyList.action", nil, func(r *resty.Request) {
		r.SetQueryParams(clientSuffix())
		r.SetResult(&resp)
	}, account)
	if err != nil {
		return nil, err
	}
	return resp.FamilyInfoResp, nil
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
		_, err = client.Request(http.MethodGet, fullUrl, nil, func(r *resty.Request) {
			r.SetQueryParams(clientSuffix()).
				SetQueryParams(map[string]string{
					"folderId":   file.Id,
					"fileType":   "0",
					"mediaAttr":  "0",
					"iconOption": "5",
					"pageNum":    fmt.Sprint(pageNum),
					"pageSize":   "130",
				})
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

		for _, folder := range resp.FileListAO.FolderList {
			files = append(files, model.File{
				Id:        fmt.Sprint(folder.ID),
				Name:      folder.Name,
				Size:      0,
				Type:      conf.FOLDER,
				Driver:    driver.Config().Name,
				UpdatedAt: MustParseTime(folder.LastOpTime),
			})
		}
		for _, file := range resp.FileListAO.FileList {
			files = append(files, model.File{
				Id:        fmt.Sprint(file.ID),
				Name:      file.Name,
				Size:      file.Size,
				Type:      utils.GetFileType(filepath.Ext(file.Name)),
				Driver:    driver.Config().Name,
				UpdatedAt: MustParseTime(file.LastOpTime),
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
	_, err = GetState(account).Request(http.MethodGet, fullUrl, nil, func(r *resty.Request) {
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

	_, err = GetState(account).Request(http.MethodPost, fullUrl, nil, func(r *resty.Request) {
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

	_, err = GetState(account).Request(http.MethodPost, API_URL+"/batch/createBatchTask.action", nil, func(r *resty.Request) {
		r.SetFormData(clientSuffix()).SetFormData(map[string]string{
			"type": "MOVE",
			"taskInfos": string(MustToBytes(utils.Json.Marshal(
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
	method := http.MethodPost
	if isFamily(account) {
		fullUrl += "/family/file"
		method = http.MethodGet
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
	method := http.MethodPost
	if isFamily(account) {
		fullUrl += "/family/file"
		method = http.MethodGet
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

	_, err = GetState(account).Request(http.MethodPost, API_URL+"/batch/createBatchTask.action", nil, func(r *resty.Request) {
		r.SetFormData(clientSuffix()).SetFormData(map[string]string{
			"type": "COPY",
			"taskInfos": string(MustToBytes(utils.Json.Marshal(
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

	_, err = GetState(account).Request(http.MethodPost, API_URL+"/batch/createBatchTask.action", nil, func(r *resty.Request) {
		r.SetFormData(clientSuffix()).SetFormData(map[string]string{
			"type": "DELETE",
			"taskInfos": string(MustToBytes(utils.Json.Marshal(
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

	if account.Bool1 {
		return driver.FastUpload(file, parentFile, account)
	}
	return driver.CommonUpload(file, parentFile, account)
	/*
		if isFamily(account) {
			return driver.uploadFamily(file, parentFile, account)
		}
		return driver.uploadPerson(file, parentFile, account)
	*/
}

func (driver Cloud189) CommonUpload(file *model.FileStream, parentFile *model.File, account *model.Account) error {
	// 初始化上传
	state := GetState(account)
	const DEFAULT int64 = 10485760
	count := int(math.Ceil(float64(file.Size) / float64(DEFAULT)))

	params := Params{
		"parentFolderId": parentFile.Id,
		"fileName":       url.PathEscape(file.Name),
		"fileSize":       fmt.Sprint(file.Size),
		"sliceSize":      fmt.Sprint(DEFAULT),
		"lazyCheck":      "1",
	}

	fullUrl := UPLOAD_URL
	if isFamily(account) {
		params.Set("familyId", account.SiteId)
		fullUrl += "/family"
	} else {
		//params.Set("extend", `{"opScene":"1","relativepath":"","rootfolderid":""}`)
		fullUrl += "/person"
	}

	var initMultiUpload InitMultiUploadResp
	_, err := state.Request(http.MethodGet, fullUrl+"/initMultiUpload", params, func(r *resty.Request) { r.SetQueryParams(clientSuffix()).SetResult(&initMultiUpload) }, account)
	if err != nil {
		return err
	}

	fileMd5 := md5.New()
	silceMd5 := md5.New()
	silceMd5Hexs := make([]string, 0, count)
	byteData := bytes.NewBuffer(make([]byte, DEFAULT))
	for i := 1; i <= count; i++ {
		byteData.Reset()
		silceMd5.Reset()
		if n, err := io.CopyN(io.MultiWriter(fileMd5, silceMd5, byteData), file, DEFAULT); err != io.EOF && n == 0 {
			return err
		}
		md5Bytes := silceMd5.Sum(nil)
		silceMd5Hexs = append(silceMd5Hexs, strings.ToUpper(hex.EncodeToString(md5Bytes)))
		silceMd5Base64 := base64.StdEncoding.EncodeToString(md5Bytes)

		var uploadUrl UploadUrlsResp
		_, err = state.Request(http.MethodGet, fullUrl+"/getMultiUploadUrls",
			Params{"partInfo": fmt.Sprintf("%d-%s", i, silceMd5Base64), "uploadFileId": initMultiUpload.Data.UploadFileID},
			func(r *resty.Request) { r.SetQueryParams(clientSuffix()).SetResult(&uploadUrl) },
			account)
		if err != nil {
			return err
		}

		uploadData := uploadUrl.UploadUrls[fmt.Sprint("partNumber_", i)]
		req, _ := http.NewRequest(http.MethodPut, uploadData.RequestURL, byteData)
		req.Header.Set("User-Agent", "")
		for k, v := range ParseHttpHeader(uploadData.RequestHeader) {
			req.Header.Set(k, v)
		}
		for k, v := range clientSuffix() {
			req.URL.RawQuery += fmt.Sprintf("&%s=%s", k, v)
		}
		r, err := base.HttpClient.Do(req)
		if err != nil {
			return err
		}
		if r.StatusCode != http.StatusOK {
			data, _ := io.ReadAll(r.Body)
			r.Body.Close()
			return fmt.Errorf(string(data))
		}
		r.Body.Close()
	}

	fileMd5Hex := strings.ToUpper(hex.EncodeToString(fileMd5.Sum(nil)))
	sliceMd5Hex := fileMd5Hex
	if int64(file.Size) > DEFAULT {
		sliceMd5Hex = strings.ToUpper(utils.GetMD5Encode(strings.Join(silceMd5Hexs, "\n")))
	}

	_, err = state.Request(http.MethodGet, fullUrl+"/commitMultiUploadFile",
		Params{
			"uploadFileId": initMultiUpload.Data.UploadFileID,
			"fileMd5":      fileMd5Hex,
			"sliceMd5":     sliceMd5Hex,
			"lazyCheck":    "1",
			"isLog":        "0",
			"opertype":     "3",
		},
		func(r *resty.Request) { r.SetQueryParams(clientSuffix()) }, account)
	return err
}

func (driver Cloud189) FastUpload(file *model.FileStream, parentFile *model.File, account *model.Account) error {
	tempFile, err := ioutil.TempFile(conf.Conf.TempDir, "file-*")
	if err != nil {
		return err
	}
	defer func() {
		_ = tempFile.Close()
		_ = os.Remove(tempFile.Name())
	}()
	// 初始化上传
	state := GetState(account)

	const DEFAULT int64 = 10485760
	count := int(math.Ceil(float64(file.Size) / float64(DEFAULT)))

	// 优先计算所需信息
	fileMd5 := md5.New()
	silceMd5 := md5.New()
	silceMd5Hexs := make([]string, 0, count)
	silceMd5Base64s := make([]string, 0, count)
	for i := 1; i <= count; i++ {
		silceMd5.Reset()
		if n, err := io.CopyN(io.MultiWriter(fileMd5, silceMd5, tempFile), file, DEFAULT); err != nil && n == 0 {
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

	params := Params{
		"parentFolderId": parentFile.Id,
		"fileName":       url.PathEscape(file.Name),
		"fileSize":       fmt.Sprint(file.Size),
		"fileMd5":        fileMd5Hex,
		"sliceSize":      fmt.Sprint(DEFAULT),
		"sliceMd5":       sliceMd5Hex,
	}

	fullUrl := UPLOAD_URL
	if isFamily(account) {
		params.Set("familyId", account.SiteId)
		fullUrl += "/family"
	} else {
		//params.Set("extend", `{"opScene":"1","relativepath":"","rootfolderid":""}`)
		fullUrl += "/person"
	}

	var uploadInfo InitMultiUploadResp
	_, err = state.Request(http.MethodGet, fullUrl+"/initMultiUpload", params, func(r *resty.Request) { r.SetQueryParams(clientSuffix()).SetResult(&uploadInfo) }, account)
	if err != nil {
		return err
	}

	if uploadInfo.Data.FileDataExists != 1 {
		var uploadUrls UploadUrlsResp
		_, err := state.Request(http.MethodGet, fullUrl+"/getMultiUploadUrls",
			Params{
				"uploadFileId": uploadInfo.Data.UploadFileID,
				"partInfo":     strings.Join(silceMd5Base64s, ","),
			},
			func(r *resty.Request) { r.SetQueryParams(clientSuffix()).SetResult(&uploadUrls) },
			account)
		if err != nil {
			return err
		}
		for i := 1; i <= count; i++ {
			uploadData := uploadUrls.UploadUrls[fmt.Sprint("partNumber_", i)]
			req, _ := http.NewRequest(http.MethodPut, uploadData.RequestURL, io.NewSectionReader(tempFile, int64(i-1)*DEFAULT, DEFAULT))
			req.Header.Set("User-Agent", "")
			for k, v := range ParseHttpHeader(uploadData.RequestHeader) {
				req.Header.Set(k, v)
			}
			for k, v := range clientSuffix() {
				req.URL.RawQuery += fmt.Sprintf("&%s=%s", k, v)
			}
			r, err := base.HttpClient.Do(req)
			if err != nil {
				return err
			}
			if r.StatusCode != http.StatusOK {
				data, _ := io.ReadAll(r.Body)
				r.Body.Close()
				return fmt.Errorf(string(data))
			}
			r.Body.Close()
		}
	}

	_, err = state.Request(http.MethodGet, fullUrl+"/commitMultiUploadFile",
		Params{
			"uploadFileId": uploadInfo.Data.UploadFileID,
			"isLog":        "0",
			"opertype":     "3",
		},
		func(r *resty.Request) { r.SetQueryParams(clientSuffix()) },
		account)
	return err
}

/*
func (driver Cloud189) uploadFamily(file *model.FileStream, parentFile *model.File, account *model.Account) error {
	tempFile, err := ioutil.TempFile(conf.Conf.TempDir, "file-*")
	if err != nil {
		return err
	}

	defer func() {
			_ = tempFile.Close()
			_ = os.Remove(tempFile.Name())
	}()

	fileMd5 := md5.New()
	if _, err = io.Copy(io.MultiWriter(fileMd5, tempFile), file); err != nil {
		return err
	}

	client := GetState(account)
	var createUpload CreateUploadFileResult
	_, err = client.Request(http.MethodGet, API_URL+"/family/file/createFamilyFile.action", nil, func(r *resty.Request) {
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
		if createUpload.UploadFileId, err = driver.uploadFileData(file, tempFile, createUpload, account); err != nil {
			return err
		}
	}

	_, err = client.Request(http.MethodGet, createUpload.FileCommitUrl, nil, func(r *resty.Request) {
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

	defer func() {
			_ = tempFile.Close()
			_ = os.Remove(tempFile.Name())
	}()

	fileMd5 := md5.New()
	if _, err = io.Copy(io.MultiWriter(fileMd5, tempFile), file); err != nil {
		return err
	}

	client := GetState(account)
	var createUpload CreateUploadFileResult
	_, err = client.Request(http.MethodPost, API_URL+"/createUploadFile.action", nil, func(r *resty.Request) {
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
		if createUpload.UploadFileId, err = driver.uploadFileData(file, tempFile, createUpload, account); err != nil {
			return err
		}
	}

	_, err = client.Request(http.MethodPost, createUpload.FileCommitUrl, nil, func(r *resty.Request) {
		r.SetQueryParams(clientSuffix())
		r.SetFormData(map[string]string{
			"uploadFileId": fmt.Sprint(createUpload.UploadFileId),
			"opertype":     "5", //5 覆盖 1 重命名
			"ResumePolicy": "1",
			"isLog":        "0",
		})
	}, account)
	return err
}

func (driver Cloud189) uploadFileData(file *model.FileStream, tempFile *os.File, createUpload CreateUploadFileResult, account *model.Account) (int64, error) {
	uploadFileState, err := driver.getUploadFileState(createUpload.UploadFileId, account)
	if err != nil {
		return 0, err
	}

	if uploadFileState.FileDataExists == 1 || uploadFileState.DataSize == int64(file.Size) {
		return uploadFileState.UploadFileId, nil
	}

	if _, err = tempFile.Seek(uploadFileState.DataSize, io.SeekStart); err != nil {
		return 0, err
	}

	_, err = GetState(account).Request("PUT", uploadFileState.FileUploadUrl, nil, func(r *resty.Request) {
		r.SetQueryParams(clientSuffix())
		r.SetHeaders(map[string]string{
			"Content-Type":           "application/octet-stream",
			"ResumePolicy":           "1",
			"Edrive-UploadFileRange": fmt.Sprintf("bytes=%d-%d", uploadFileState.DataSize, file.Size),
			"Expect":                 "100-continue",
		})
		if isFamily(account) {
			r.SetHeaders(map[string]string{
				"familyId":     account.SiteId,
				"UploadFileId": fmt.Sprint(uploadFileState.UploadFileId),
			})
		} else {
			r.SetHeader("Edrive-UploadFileId", fmt.Sprint(uploadFileState.UploadFileId))
		}
		r.SetBody(tempFile)
	}, account)
	return uploadFileState.UploadFileId, err
}

func (driver Cloud189) getUploadFileState(uploadFileId int64, account *model.Account) (*UploadFileStatusResult, error) {
	fullUrl := API_URL
	if isFamily(account) {
		fullUrl += "/family/file/getFamilyFileStatus.action"
	} else {
		fullUrl += "/getUploadFileStatus.action"
	}
	var uploadFileState UploadFileStatusResult
	_, err := GetState(account).Request(http.MethodGet, fullUrl, nil, func(r *resty.Request) {
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
}*/

var _ base.Driver = (*Cloud189)(nil)
