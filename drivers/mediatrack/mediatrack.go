package mediatrack

import (
	"errors"
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"path"
	"strconv"
	"time"
)

type BaseResp struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type T struct {
	BaseResp
}

func (driver MediaTrack) Request(url string, method int, headers, query, form map[string]string, data interface{}, resp interface{}, account *model.Account) ([]byte, error) {
	req := base.RestyClient.R()
	req.SetHeader("Authorization", "Bearer "+account.AccessToken)
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
	var e BaseResp
	req.SetResult(&e)
	var err error
	var res *resty.Response
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
	log.Debugln(res.String())
	if e.Status != "SUCCESS" {
		return nil, errors.New(e.Message)
	}
	if resp != nil {
		err = utils.Json.Unmarshal(res.Body(), resp)
	}
	return res.Body(), err
}

type File struct {
	Category     int           `json:"category"`
	ChildAssets  []interface{} `json:"childAssets"`
	CommentCount int           `json:"comment_count"`
	CoverAsset   interface{}   `json:"cover_asset"`
	CoverAssetID string        `json:"cover_asset_id"`
	CreatedAt    time.Time     `json:"created_at"`
	DeletedAt    string        `json:"deleted_at"`
	Description  string        `json:"description"`
	File         *struct {
		Cover string `json:"cover"`
		Src   string `json:"src"`
	} `json:"file"`
	//FileID string `json:"file_id"`
	ID string `json:"id"`

	Size       string        `json:"size"`
	Thumbnails []interface{} `json:"thumbnails"`
	Title      string        `json:"title"`
	UpdatedAt  time.Time     `json:"updated_at"`
}

func (driver MediaTrack) formatFile(file *File) *model.File {
	f := model.File{
		Id:        file.ID,
		Name:      file.Title,
		Size:      0,
		Driver:    driver.Config().Name,
		UpdatedAt: &file.UpdatedAt,
	}
	if file.File == nil {
		// folder
		f.Type = conf.FOLDER
	} else {
		size, _ := strconv.ParseInt(file.Size, 10, 64)
		f.Size = size
		f.Type = utils.GetFileType(path.Ext(file.Title))
		if file.File.Cover != "" {
			f.Thumbnail = "https://nano.mtres.cn/" + file.File.Cover
		}
	}
	return &f
}

type ChildrenResp struct {
	Status string `json:"status"`
	Data   struct {
		Total  int    `json:"total"`
		Assets []File `json:"assets"`
	} `json:"data"`
	Path      string `json:"path"`
	TraceID   string `json:"trace_id"`
	RequestID string `json:"requestId"`
}

func (driver MediaTrack) GetFiles(parentId string, account *model.Account) ([]model.File, error) {
	files := make([]model.File, 0)
	url := fmt.Sprintf("https://jayce.api.mediatrack.cn/v4/assets/%s/children", parentId)
	sort := ""
	if account.OrderBy != "" {
		if account.OrderDirection == "true" {
			sort = "-"
		}
		sort += account.OrderBy
	}
	page := 1
	for {
		var resp ChildrenResp
		_, err := driver.Request(url, base.Get, nil, map[string]string{
			"page": strconv.Itoa(page),
			"size": "50",
			"sort": sort,
		}, nil, nil, &resp, account)
		if err != nil {
			return nil, err
		}
		if len(resp.Data.Assets) == 0 {
			break
		}
		page++
		for _, file := range resp.Data.Assets {
			files = append(files, *driver.formatFile(&file))
		}
	}
	return files, nil
}

type UploadResp struct {
	Status string `json:"status"`
	Data   struct {
		Credentials struct {
			TmpSecretID  string    `json:"TmpSecretId"`
			TmpSecretKey string    `json:"TmpSecretKey"`
			Token        string    `json:"Token"`
			ExpiredTime  int       `json:"ExpiredTime"`
			Expiration   time.Time `json:"Expiration"`
			StartTime    int       `json:"StartTime"`
		} `json:"credentials"`
		Object string `json:"object"`
		Bucket string `json:"bucket"`
		Region string `json:"region"`
		URL    string `json:"url"`
		Size   string `json:"size"`
	} `json:"data"`
	Path      string `json:"path"`
	TraceID   string `json:"trace_id"`
	RequestID string `json:"requestId"`
}

func init() {
	base.RegisterDriver(&MediaTrack{})
}
