package teambition

import (
	"errors"
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"io"
	"path"
	"strconv"
	"time"
)

type ErrResp struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

func (driver Teambition) Request(pathname string, method int, headers, query, form map[string]string, data interface{}, resp interface{}, account *model.Account) ([]byte, error) {
	url := "https://www.teambition.com" + pathname
	if account.InternalType == "International" {
		url = "https://us.teambition.com" + pathname
	}
	req := base.RestyClient.R()
	req.SetHeader("Cookie", account.AccessToken)
	if headers != nil {
		req.SetHeaders(headers)
	}
	if query != nil {
		req.SetQueryParams(query)
	}
	if form != nil {
		req.SetFormData(form)
	}
	if data != nil {
		req.SetBody(data)
	}
	if resp != nil {
		req.SetResult(resp)
	}
	var e ErrResp
	var err error
	var res *resty.Response
	req.SetError(&e)
	switch method {
	case base.Get:
		res, err = req.Get(url)
	case base.Post:
		res, err = req.Post(url)
	case base.Delete:
		res, err = req.Delete(url)
	case base.Patch:
		res, err = req.Patch(url)
	case base.Put:
		res, err = req.Put(url)
	default:
		return nil, base.ErrNotSupport
	}
	if err != nil {
		return nil, err
	}
	if e.Name != "" {
		return nil, errors.New(e.Message)
	}
	return res.Body(), nil
}

func (driver Teambition) GetFiles(parentId string, account *model.Account) ([]model.File, error) {
	files := make([]model.File, 0)
	page := 1
	for {
		var collections []Collection
		_, err := driver.Request("/api/collections", base.Get, nil, map[string]string{
			"_parentId":  parentId,
			"_projectId": account.Zone,
			"order":      account.OrderBy + account.OrderDirection,
			"count":      "50",
			"page":       strconv.Itoa(page),
		}, nil, nil, &collections, account)
		if err != nil {
			return nil, err
		}
		if len(collections) == 0 {
			break
		}
		page++
		for _, collection := range collections {
			if collection.Title == "" {
				continue
			}
			files = append(files, model.File{
				Id:        collection.ID,
				Name:      collection.Title,
				Size:      0,
				Type:      conf.FOLDER,
				Driver:    driver.Config().Name,
				UpdatedAt: collection.Updated,
			})
		}
	}
	page = 1
	for {
		var works []Work
		_, err := driver.Request("/api/works", base.Get, nil, map[string]string{
			"_parentId":  parentId,
			"_projectId": account.Zone,
			"order":      account.OrderBy + account.OrderDirection,
			"count":      "50",
			"page":       strconv.Itoa(page),
		}, nil, nil, &works, account)
		if err != nil {
			return nil, err
		}
		if len(works) == 0 {
			break
		}
		page++
		for _, work := range works {
			files = append(files, model.File{
				Id:        work.ID,
				Name:      work.FileName,
				Size:      work.FileSize,
				Type:      utils.GetFileType(path.Ext(work.FileName)),
				Driver:    driver.Config().Name,
				UpdatedAt: work.Updated,
				Thumbnail: work.Thumbnail,
				Url:       work.DownloadURL,
			})
		}
	}
	return files, nil
}

func (driver Teambition) upload(file *model.FileStream, token string, account *model.Account) (*FileUpload, error) {
	prefix := "tcs"
	if account.InternalType == "International" {
		prefix = "us-tcs"
	}
	var newFile FileUpload
	_, err := base.RestyClient.R().SetResult(&newFile).SetHeader("Authorization", token).
		SetMultipartFormData(map[string]string{
			"name": file.GetFileName(),
			"type": file.GetMIMEType(),
			"size": strconv.FormatUint(file.GetSize(), 10),
			//"lastModifiedDate": "",
		}).SetMultipartField("file", file.GetFileName(), file.GetMIMEType(), file).
		Post(fmt.Sprintf("https://%s.teambition.net/upload", prefix))
	if err != nil {
		return nil, err
	}
	return &newFile, nil
}

func (driver Teambition) chunkUpload(file *model.FileStream, token string, account *model.Account) (*FileUpload, error) {
	prefix := "tcs"
	referer := "https://www.teambition.com/"
	if account.InternalType == "International" {
		prefix = "us-tcs"
		referer = "https://us.teambition.com/"
	}
	var newChunk ChunkUpload
	_, err := base.RestyClient.R().SetResult(&newChunk).SetHeader("Authorization", token).
		SetBody(base.Json{
			"fileName":    file.GetFileName(),
			"fileSize":    file.GetSize(),
			"lastUpdated": time.Now(),
		}).Post(fmt.Sprintf("https://%s.teambition.net/upload/chunk", prefix))
	if err != nil {
		return nil, err
	}
	for i := 0; i < newChunk.Chunks; i++ {
		chunkSize := newChunk.ChunkSize
		if i == newChunk.Chunks-1 {
			chunkSize = int(file.GetSize()) - i*chunkSize
		}
		log.Debugf("%d : %d", i, chunkSize)
		chunkData := make([]byte, chunkSize)
		_, err = io.ReadFull(file, chunkData)
		if err != nil {
			return nil, err
		}
		u := fmt.Sprintf("https://%s.teambition.net/upload/chunk/%s?chunk=%d&chunks=%d",
			prefix, newChunk.FileKey, i+1, newChunk.Chunks)
		log.Debugf("url: %s", u)
		res, err := base.RestyClient.R().SetHeaders(map[string]string{
			"Authorization": token,
			"Content-Type":  "application/octet-stream",
			"Referer":       referer,
		}).SetBody(chunkData).Post(u)
		if err != nil {
			return nil, err
		}
		log.Debug(res.Status(), res.String())
		//req, err := http.NewRequest("POST",
		//	u,
		//	bytes.NewBuffer(chunkData))
		//if err != nil {
		//	return nil, err
		//}
		//req.Header.Set("Authorization", token)
		//req.Header.Set("Content-Type", "application/octet-stream")
		//req.Header.Set("Referer", "https://www.teambition.com/")
		//resp, err := base.HttpClient.Do(req)
		//res, _ := ioutil.ReadAll(resp.Body)
		//log.Debugf("chunk upload status: %s, res: %s", resp.Status, string(res))
		if err != nil {
			return nil, err
		}
	}
	res, err := base.RestyClient.R().SetHeader("Authorization", token).Post(
		fmt.Sprintf("https://%s.teambition.net/upload/chunk/%s",
			prefix, newChunk.FileKey))
	log.Debug(res.Status(), res.String())
	if err != nil {
		return nil, err
	}
	return &newChunk.FileUpload, nil
}

func (driver Teambition) finishUpload(file *FileUpload, parentId string, account *model.Account) error {
	file.InvolveMembers = []interface{}{}
	file.Visible = "members"
	file.ParentId = parentId
	_, err := driver.Request("/api/works", base.Post, nil, nil, nil, base.Json{
		"works":     []FileUpload{*file},
		"_parentId": parentId,
	}, nil, account)
	return err
}

func init() {
	base.RegisterDriver(&Teambition{})
}
