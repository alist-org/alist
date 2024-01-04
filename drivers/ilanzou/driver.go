package template

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/foxxorcat/mopan-sdk-go"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

type ILanZou struct {
	model.Storage
	Addition

	userID string
}

func (d *ILanZou) Config() driver.Config {
	return config
}

func (d *ILanZou) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *ILanZou) Init(ctx context.Context) error {
	if d.UUID == "" {
		res, err := d.unproved("/getUuid", http.MethodGet, nil)
		if err != nil {
			return err
		}
		d.UUID = utils.Json.Get(res, "uuid").ToString()
	}
	res, err := d.proved("/user/account/map", http.MethodGet, nil)
	if err != nil {
		return err
	}
	d.userID = utils.Json.Get(res, "map", "userId").ToString()
	log.Debugf("[ilanzou] init response: %s", res)
	return nil
}

func (d *ILanZou) Drop(ctx context.Context) error {
	return nil
}

func (d *ILanZou) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	offset := 1
	limit := 60
	var res []ListItem
	for {
		var resp ListResp
		_, err := d.proved("/record/file/list", http.MethodGet, func(req *resty.Request) {
			req.SetQueryParams(map[string]string{
				"type":     "0",
				"folderId": dir.GetID(),
				"offset":   strconv.Itoa(offset),
				"limit":    strconv.Itoa(limit),
			}).SetResult(&resp)
		})
		if err != nil {
			return nil, err
		}
		res = append(res, resp.List...)
		if resp.TotalPage <= resp.Offset {
			break
		}
		offset++
	}
	return utils.SliceConvert(res, func(f ListItem) (model.Obj, error) {
		updTime, err := time.ParseInLocation("2006-01-02 15:04:05", f.UpdTime, time.Local)
		if err != nil {
			return nil, err
		}
		obj := model.Object{
			ID: strconv.FormatInt(f.FileId, 10),
			//Path:     "",
			Name:     f.FileName,
			Size:     f.FileSize * 1024,
			Modified: updTime,
			Ctime:    updTime,
			IsFolder: false,
			//HashInfo: utils.HashInfo{},
		}
		if f.FileType == 2 {
			obj.IsFolder = true
			obj.Size = 0
			obj.ID = strconv.FormatInt(f.FolderId, 10)
			obj.Name = f.FolderName
		}
		return &obj, nil
	})
}

func (d *ILanZou) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	u, err := url.Parse("https://api.ilanzou.com/unproved/file/redirect")
	if err != nil {
		return nil, err
	}
	query := u.Query()
	query.Set("uuid", d.UUID)
	query.Set("devType", "6")
	query.Set("devCode", d.UUID)
	query.Set("devModel", "chrome")
	query.Set("devVersion", "120")
	query.Set("appVersion", "")
	ts, err := getTimestamp()
	if err != nil {
		return nil, err
	}
	query.Set("timestamp", ts)
	//query.Set("appToken", d.Token)
	query.Set("enable", "1")
	downloadId, err := mopan.AesEncrypt([]byte(fmt.Sprintf("%s|%s", file.GetID(), d.userID)), AesSecret)
	if err != nil {
		return nil, err
	}
	query.Set("downloadId", hex.EncodeToString(downloadId))
	auth, err := mopan.AesEncrypt([]byte(fmt.Sprintf("%s|%d", file.GetID(), time.Now().UnixMilli())), AesSecret)
	if err != nil {
		return nil, err
	}
	query.Set("auth", hex.EncodeToString(auth))
	u.RawQuery = query.Encode()
	link := model.Link{URL: u.String()}
	return &link, nil
}

func (d *ILanZou) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) (model.Obj, error) {
	res, err := d.proved("/file/folder/save", http.MethodPost, func(req *resty.Request) {
		req.SetBody(base.Json{
			"folderDesc": "",
			"folderId":   parentDir.GetID(),
			"folderName": dirName,
		})
	})
	if err != nil {
		return nil, err
	}
	return &model.Object{
		ID: utils.Json.Get(res, "list", "0", "id").ToString(),
		//Path:     "",
		Name:     dirName,
		Size:     0,
		Modified: time.Now(),
		Ctime:    time.Now(),
		IsFolder: true,
		//HashInfo: utils.HashInfo{},
	}, nil
}

func (d *ILanZou) Move(ctx context.Context, srcObj, dstDir model.Obj) (model.Obj, error) {
	var fileIds, folderIds []string
	if srcObj.IsDir() {
		folderIds = []string{srcObj.GetID()}
	} else {
		fileIds = []string{srcObj.GetID()}
	}
	_, err := d.proved("/file/folder/move", http.MethodPost, func(req *resty.Request) {
		req.SetBody(base.Json{
			"folderIds": strings.Join(folderIds, ","),
			"fileIds":   strings.Join(fileIds, ","),
			"targetId":  dstDir.GetID(),
		})
	})
	if err != nil {
		return nil, err
	}
	return srcObj, nil
}

func (d *ILanZou) Rename(ctx context.Context, srcObj model.Obj, newName string) (model.Obj, error) {
	var err error
	if srcObj.IsDir() {
		_, err = d.proved("/file/folder/edit", http.MethodPost, func(req *resty.Request) {
			req.SetBody(base.Json{
				"folderDesc": "",
				"folderId":   srcObj.GetID(),
				"folderName": newName,
			})
		})
	} else {
		_, err = d.proved("/file/edit", http.MethodPost, func(req *resty.Request) {
			req.SetBody(base.Json{
				"fileDesc": "",
				"fileId":   srcObj.GetID(),
				"fileName": newName,
			})
		})
	}
	if err != nil {
		return nil, err
	}
	return &model.Object{
		ID: srcObj.GetID(),
		//Path:     "",
		Name:     newName,
		Size:     srcObj.GetSize(),
		Modified: time.Now(),
		Ctime:    srcObj.CreateTime(),
		IsFolder: srcObj.IsDir(),
	}, nil
}

func (d *ILanZou) Copy(ctx context.Context, srcObj, dstDir model.Obj) (model.Obj, error) {
	// TODO copy obj, optional
	return nil, errs.NotImplement
}

func (d *ILanZou) Remove(ctx context.Context, obj model.Obj) error {
	var fileIds, folderIds []string
	if obj.IsDir() {
		folderIds = []string{obj.GetID()}
	} else {
		fileIds = []string{obj.GetID()}
	}
	_, err := d.proved("/file/delete", http.MethodPost, func(req *resty.Request) {
		req.SetBody(base.Json{
			"folderIds": strings.Join(folderIds, ","),
			"fileIds":   strings.Join(fileIds, ","),
			"status":    0,
		})
	})
	return err
}

const DefaultPartSize = 1024 * 1024 * 8

func (d *ILanZou) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) (model.Obj, error) {
	return nil, errs.NotImplement
}

//func (d *ILanZou) Other(ctx context.Context, args model.OtherArgs) (interface{}, error) {
//	return nil, errs.NotSupport
//}

var _ driver.Driver = (*ILanZou)(nil)
