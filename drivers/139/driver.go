package _39

import (
	"bytes"
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	log "github.com/sirupsen/logrus"
	"io"
	"math"
	"net/http"
	"path/filepath"
	"strconv"
)

type Cloud139 struct{}

func (driver Cloud139) Config() base.DriverConfig {
	return base.DriverConfig{
		Name:      "139Yun",
		LocalSort: true,
	}
}

func (driver Cloud139) Items() []base.Item {
	return []base.Item{
		{
			Name:        "username",
			Label:       "phone",
			Type:        base.TypeString,
			Required:    true,
			Description: "phone number",
		},
		{
			Name:        "access_token",
			Label:       "Cookie",
			Type:        base.TypeString,
			Required:    true,
			Description: "Unknown expiration time",
		},
		{
			Name:     "internal_type",
			Label:    "139yun type",
			Type:     base.TypeSelect,
			Required: true,
			Values:   "Personal,Family",
		},
		{
			Name:     "root_folder",
			Label:    "root folder file_id",
			Type:     base.TypeString,
			Required: true,
		},
		{
			Name:     "site_id",
			Label:    "cloud_id",
			Type:     base.TypeString,
			Required: false,
		},
	}
}

func (driver Cloud139) Save(account *model.Account, old *model.Account) error {
	if account == nil {
		return nil
	}
	_, err := driver.Request("/orchestration/personalCloud/user/v1.0/qryUserExternInfo", base.Post, nil, nil, nil, base.Json{
		"qryUserExternInfoReq": base.Json{
			"commonAccountInfo": base.Json{
				"account":     account.Username,
				"accountType": 1,
			},
		},
	}, nil, account)
	return err
}

func (driver Cloud139) File(path string, account *model.Account) (*model.File, error) {
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

func (driver Cloud139) Files(path string, account *model.Account) ([]model.File, error) {
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
		if isFamily(account) {
			files, err = driver.familyGetFiles(file.Id, account)
		} else {
			files, err = driver.GetFiles(file.Id, account)
		}
		if err != nil {
			return nil, err
		}
		if len(files) > 0 {
			_ = base.SetCache(path, files, account)
		}
	}
	return files, nil
}

func (driver Cloud139) Link(args base.Args, account *model.Account) (*base.Link, error) {
	file, err := driver.File(args.Path, account)
	if err != nil {
		return nil, err
	}
	var u string
	//if isFamily(account) {
	//	u, err = driver.familyLink(file.Id, account)
	//} else {
	u, err = driver.GetLink(file.Id, account)
	//}
	if err != nil {
		return nil, err
	}
	return &base.Link{Url: u}, nil
}

func (driver Cloud139) Path(path string, account *model.Account) (*model.File, []model.File, error) {
	path = utils.ParsePath(path)
	log.Debugf("139 path: %s", path)
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

//func (driver Cloud139) Proxy(r *http.Request, account *model.Account) {
//
//}

func (driver Cloud139) Preview(path string, account *model.Account) (interface{}, error) {
	return nil, base.ErrNotSupport
}

func (driver Cloud139) MakeDir(path string, account *model.Account) error {
	parentFile, err := driver.File(utils.Dir(path), account)
	if err != nil {
		return err
	}
	data := base.Json{
		"createCatalogExtReq": base.Json{
			"parentCatalogID": parentFile.Id,
			"newCatalogName":  utils.Base(path),
			"commonAccountInfo": base.Json{
				"account":     account.Username,
				"accountType": 1,
			},
		},
	}
	pathname := "/orchestration/personalCloud/catalog/v1.0/createCatalogExt"
	if isFamily(account) {
		data = base.Json{
			"cloudID": account.SiteId,
			"commonAccountInfo": base.Json{
				"account":     account.Username,
				"accountType": 1,
			},
			"docLibName": utils.Base(path),
		}
		pathname = "/orchestration/familyCloud/cloudCatalog/v1.0/createCloudDoc"
	}
	_, err = driver.Post(pathname,
		data, nil, account)
	return err
}

func (driver Cloud139) Move(src string, dst string, account *model.Account) error {
	if isFamily(account) {
		return base.ErrNotSupport
	}
	srcFile, err := driver.File(src, account)
	if err != nil {
		return err
	}
	dstParentFile, err := driver.File(utils.Dir(dst), account)
	if err != nil {
		return err
	}
	var contentInfoList []string
	var catalogInfoList []string
	if srcFile.IsDir() {
		catalogInfoList = append(catalogInfoList, srcFile.Id)
	} else {
		contentInfoList = append(contentInfoList, srcFile.Id)
	}
	data := base.Json{
		"createBatchOprTaskReq": base.Json{
			"taskType":   3,
			"actionType": "304",
			"taskInfo": base.Json{
				"contentInfoList": contentInfoList,
				"catalogInfoList": catalogInfoList,
				"newCatalogID":    dstParentFile.Id,
			},
			"commonAccountInfo": base.Json{
				"account":     account.Username,
				"accountType": 1,
			},
		},
	}
	pathname := "/orchestration/personalCloud/batchOprTask/v1.0/createBatchOprTask"
	_, err = driver.Post(pathname, data, nil, account)
	return err
}

func (driver Cloud139) Rename(src string, dst string, account *model.Account) error {
	if isFamily(account) {
		return base.ErrNotSupport
	}
	srcFile, err := driver.File(src, account)
	if err != nil {
		return err
	}
	var data base.Json
	var pathname string
	if srcFile.IsDir() {
		data = base.Json{
			"catalogID":   srcFile.Id,
			"catalogName": utils.Base(dst),
			"commonAccountInfo": base.Json{
				"account":     account.Username,
				"accountType": 1,
			},
		}
		pathname = "/orchestration/personalCloud/catalog/v1.0/updateCatalogInfo"
	} else {
		data = base.Json{
			"contentID":   srcFile.Id,
			"contentName": utils.Base(dst),
			"commonAccountInfo": base.Json{
				"account":     account.Username,
				"accountType": 1,
			},
		}
		pathname = "/orchestration/personalCloud/content/v1.0/updateContentInfo"
	}
	_, err = driver.Post(pathname, data, nil, account)
	return err
}

func (driver Cloud139) Copy(src string, dst string, account *model.Account) error {
	if isFamily(account) {
		return base.ErrNotSupport
	}
	srcFile, err := driver.File(src, account)
	if err != nil {
		return err
	}
	dstParentFile, err := driver.File(utils.Dir(dst), account)
	if err != nil {
		return err
	}
	var contentInfoList []string
	var catalogInfoList []string
	if srcFile.IsDir() {
		catalogInfoList = append(catalogInfoList, srcFile.Id)
	} else {
		contentInfoList = append(contentInfoList, srcFile.Id)
	}
	data := base.Json{
		"createBatchOprTaskReq": base.Json{
			"taskType":   3,
			"actionType": 309,
			"taskInfo": base.Json{
				"contentInfoList": contentInfoList,
				"catalogInfoList": catalogInfoList,
				"newCatalogID":    dstParentFile.Id,
			},
			"commonAccountInfo": base.Json{
				"account":     account.Username,
				"accountType": 1,
			},
		},
	}
	pathname := "/orchestration/personalCloud/batchOprTask/v1.0/createBatchOprTask"
	_, err = driver.Post(pathname, data, nil, account)
	return err
}

func (driver Cloud139) Delete(path string, account *model.Account) error {
	file, err := driver.File(path, account)
	if err != nil {
		return err
	}
	var contentInfoList []string
	var catalogInfoList []string
	if file.IsDir() {
		catalogInfoList = append(catalogInfoList, file.Id)
	} else {
		contentInfoList = append(contentInfoList, file.Id)
	}
	data := base.Json{
		"createBatchOprTaskReq": base.Json{
			"taskType":   2,
			"actionType": 201,
			"taskInfo": base.Json{
				"newCatalogID":    "",
				"contentInfoList": contentInfoList,
				"catalogInfoList": catalogInfoList,
			},
			"commonAccountInfo": base.Json{
				"account":     account.Username,
				"accountType": 1,
			},
		},
	}
	pathname := "/orchestration/personalCloud/batchOprTask/v1.0/createBatchOprTask"
	if isFamily(account) {
		data = base.Json{
			"catalogList": catalogInfoList,
			"contentList": contentInfoList,
			"commonAccountInfo": base.Json{
				"account":     account.Username,
				"accountType": 1,
			},
			"sourceCatalogType": 1002,
			"taskType":          2,
		}
		pathname = "/orchestration/familyCloud/batchOprTask/v1.0/createBatchOprTask"
	}
	_, err = driver.Post(pathname, data, nil, account)
	return err
}

func (driver Cloud139) Upload(file *model.FileStream, account *model.Account) error {
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
	data := base.Json{
		"manualRename": 2,
		"operation":    0,
		"fileCount":    1,
		"totalSize":    file.Size,
		"uploadContentList": []base.Json{{
			"contentName": file.Name,
			"contentSize": file.Size,
			// "digest": "5a3231986ce7a6b46e408612d385bafa"
		}},
		"parentCatalogID": parentFile.Id,
		"newCatalogName":  "",
		"commonAccountInfo": base.Json{
			"account":     account.Username,
			"accountType": 1,
		},
	}
	pathname := "/orchestration/personalCloud/uploadAndDownload/v1.0/pcUploadFileRequest"
	if isFamily(account) {
		data = newJson(base.Json{
			"fileCount":    1,
			"manualRename": 2,
			"operation":    0,
			"path":         "",
			"seqNo":        "",
			"totalSize":    file.Size,
			"uploadContentList": []base.Json{{
				"contentName": file.Name,
				"contentSize": file.Size,
				// "digest": "5a3231986ce7a6b46e408612d385bafa"
			}},
		}, account)
		pathname = "/orchestration/familyCloud/content/v1.0/getFileUploadURL"
		return base.ErrNotSupport
	}
	var resp UploadResp
	_, err = driver.Post(pathname, data, &resp, account)
	if err != nil {
		return err
	}
	var Default uint64 = 10485760
	part := int(math.Ceil(float64(file.Size) / float64(Default)))
	var start uint64 = 0
	for i := 0; i < part; i++ {
		byteSize := file.Size - start
		if byteSize > Default {
			byteSize = Default
		}
		byteData := make([]byte, byteSize)
		_, err = io.ReadFull(file, byteData)
		if err != nil {
			return err
		}
		req, err := http.NewRequest("POST", resp.Data.UploadResult.RedirectionURL, bytes.NewBuffer(byteData))
		if err != nil {
			return err
		}
		headers := map[string]string{
			"Accept":         "*/*",
			"Content-Type":   "text/plain;name=" + unicode(file.Name),
			"contentSize":    strconv.FormatUint(file.Size, 10),
			"range":          fmt.Sprintf("bytes=%d-%d", start, start+byteSize-1),
			"content-length": strconv.FormatUint(byteSize, 10),
			"uploadtaskID":   resp.Data.UploadResult.UploadTaskID,
			"rangeType":      "0",
			"Referer":        "https://yun.139.com/",
			"User-Agent":     "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/95.0.4638.69 Safari/537.36 Edg/95.0.1020.44",
			"x-SvcType":      "1",
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		res, err := base.HttpClient.Do(req)
		if err != nil {
			return err
		}
		log.Debugf("%+v", res)
		res.Body.Close()
		start += byteSize
	}
	return nil
}

var _ base.Driver = (*Cloud139)(nil)
