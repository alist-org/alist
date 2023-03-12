package baidu_share

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"

	"github.com/Xhofe/go-cache"
	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
)

type BaiduShare struct {
	model.Storage
	Addition
	config     map[string]string
	dlinkCache cache.ICache[string]
}

func (d *BaiduShare) Config() driver.Config {
	return config
}

func (d *BaiduShare) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *BaiduShare) Init(ctx context.Context) error {
	// TODO login / refresh token
	//op.MustSaveDriverStorage(d)
	d.config = map[string]string{}
	d.dlinkCache = cache.NewMemCache(cache.WithClearInterval[string](time.Duration(d.CacheExpiration) * time.Minute))
	return nil
}

func (d *BaiduShare) Drop(ctx context.Context) error {
	d.dlinkCache.Clear()
	return nil
}

func (d *BaiduShare) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	// TODO return the files list
	body := url.Values{
		"shorturl": {d.Surl},
		"dir":      {dir.GetPath()},
		"root":     {"0"},
		"pwd":      {d.Pwd},
		"num":      {"1000"},
		"order":    {"time"},
	}
	if body.Get("dir") == "" || body.Get("dir") == d.config["root"] {
		body.Set("root", "1")
	}
	res := []model.Obj{}
	var err error
	var page int64 = 1
	more := true
	for more {
		body.Set("page", strconv.FormatInt(page, 10))
		req := base.RestyClient.R().
			SetCookies([]*http.Cookie{{Name: "BDUSS", Value: d.BDUSS}}).
			SetBody(body.Encode())
		resp, e := req.Post("https://pan.baidu.com/share/wxlist?channel=weixin&version=2.2.2&clienttype=25&web=1")
		err = e
		jsonresp := jsonResp{}
		if err == nil {
			err = base.RestyClient.JSONUnmarshal(resp.Body(), &jsonresp)
		}
		if err == nil && jsonresp.Errno == 0 {
			more = jsonresp.Data.More
			page += 1
			for _, v := range jsonresp.Data.List {
				size, _ := v.Size.Int64()
				mtime, _ := v.Time.Int64()
				res = append(res, &model.Object{
					ID:       v.ID.String(),
					Path:     v.Path,
					Name:     v.Name,
					Size:     size,
					Modified: time.Unix(mtime, 0),
					IsFolder: v.Dir.String() == "1",
				})
				d.dlinkCache.Set(v.Path, v.Dlink, cache.WithEx[string](time.Duration(d.CacheExpiration/2)*time.Minute))
			}
			if len(res) > 0 && body.Get("root") == "1" {
				d.config["root"] = path.Dir(res[0].GetPath())
			}
		} else {
			if err == nil {
				err = fmt.Errorf("errno:%d", jsonresp.Errno)
			}
			break
		}
	}
	return res, err
}

func (d *BaiduShare) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	// TODO return link of file
	var err error
	req := base.RestyClient.R().SetCookies([]*http.Cookie{{Name: "BDUSS", Value: d.BDUSS}})
	url, found := d.dlinkCache.Get(file.GetPath())
	if !found {
		_, err = d.List(ctx, &model.Object{Path: path.Dir(file.GetPath())}, model.ListArgs{})
		url, found = d.dlinkCache.Get(file.GetPath())
		if err == nil && !found {
			err = errs.NotSupport
		}
	}
	return &model.Link{URL: url, Header: req.Header}, err
}

func (d *BaiduShare) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {
	// TODO create folder
	return errs.NotSupport
}

func (d *BaiduShare) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	// TODO move obj
	return errs.NotSupport
}

func (d *BaiduShare) Rename(ctx context.Context, srcObj model.Obj, newName string) error {
	// TODO rename obj
	return errs.NotSupport
}

func (d *BaiduShare) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	// TODO copy obj
	return errs.NotSupport
}

func (d *BaiduShare) Remove(ctx context.Context, obj model.Obj) error {
	// TODO remove obj
	return errs.NotSupport
}

func (d *BaiduShare) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	// TODO upload file
	return errs.NotSupport
}

//func (d *Template) Other(ctx context.Context, args model.OtherArgs) (interface{}, error) {
//	return nil, errs.NotSupport
//}

var _ driver.Driver = (*BaiduShare)(nil)
