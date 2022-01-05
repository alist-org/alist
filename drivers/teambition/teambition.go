package teambition

import (
	"errors"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/go-resty/resty/v2"
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

type Collection struct {
	ID      string    `json:"_id"`
	Title   string    `json:"title"`
	Updated time.Time `json:"updated"`
}

type Work struct {
	ID           string    `json:"_id"`
	FileName     string    `json:"fileName"`
	FileSize     int64     `json:"fileSize"`
	FileKey      string    `json:"fileKey"`
	FileCategory string    `json:"fileCategory"`
	DownloadURL  string    `json:"downloadUrl"`
	ThumbnailURL string    `json:"thumbnailUrl"`
	Thumbnail    string    `json:"thumbnail"`
	Updated      time.Time `json:"updated"`
	PreviewURL   string    `json:"previewUrl"`
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
			files = append(files, model.File{
				Id:        collection.ID,
				Name:      collection.Title,
				Size:      0,
				Type:      conf.FOLDER,
				Driver:    driver.Config().Name,
				UpdatedAt: &collection.Updated,
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
				UpdatedAt: &work.Updated,
				Thumbnail: work.Thumbnail,
				Url:       work.DownloadURL,
			})
		}
	}
	return files, nil
}

func init() {
	base.RegisterDriver(&Teambition{})
}
