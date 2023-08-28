package alist_v3

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/go-resty/resty/v2"
)

type AListV3 struct {
	model.Storage
	Addition
}

func (d *AListV3) Config() driver.Config {
	return config
}

func (d *AListV3) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *AListV3) Init(ctx context.Context) error {
	d.Addition.Address = strings.TrimSuffix(d.Addition.Address, "/")
	var resp common.Resp[MeResp]
	_, err := d.request("/me", http.MethodGet, func(req *resty.Request) {
		req.SetResult(&resp)
	})
	if err != nil {
		return err
	}
	// if the username is not empty and the username is not the same as the current username, then login again
	if d.Username != "" && d.Username != resp.Data.Username {
		err = d.login()
		if err != nil {
			return err
		}
	}
	// re-get the user info
	_, err = d.request("/me", http.MethodGet, func(req *resty.Request) {
		req.SetResult(&resp)
	})
	if err != nil {
		return err
	}
	if resp.Data.Role == model.GUEST {
		url := d.Address + "/api/public/settings"
		res, err := base.RestyClient.R().Get(url)
		if err != nil {
			return err
		}
		allowMounted := utils.Json.Get(res.Body(), "data", conf.AllowMounted).ToString() == "true"
		if !allowMounted {
			return fmt.Errorf("the site does not allow mounted")
		}
	}
	return err
}

func (d *AListV3) Drop(ctx context.Context) error {
	return nil
}

func (d *AListV3) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	var resp common.Resp[FsListResp]
	_, err := d.request("/fs/list", http.MethodPost, func(req *resty.Request) {
		req.SetResult(&resp).SetBody(ListReq{
			PageReq: model.PageReq{
				Page:    1,
				PerPage: 0,
			},
			Path:     dir.GetPath(),
			Password: d.MetaPassword,
			Refresh:  false,
		})
	})
	if err != nil {
		return nil, err
	}
	var files []model.Obj
	for _, f := range resp.Data.Content {
		file := model.ObjThumb{
			Object: model.Object{
				Name:     f.Name,
				Modified: f.Modified,
				Ctime:    f.Created,
				Size:     f.Size,
				IsFolder: f.IsDir,
				HashInfo: utils.FromString(f.HashInfo),
			},
			Thumbnail: model.Thumbnail{Thumbnail: f.Thumb},
		}
		files = append(files, &file)
	}
	return files, nil
}

func (d *AListV3) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	var resp common.Resp[FsGetResp]
	_, err := d.request("/fs/get", http.MethodPost, func(req *resty.Request) {
		req.SetResult(&resp).SetBody(FsGetReq{
			Path:     file.GetPath(),
			Password: d.MetaPassword,
		})
	})
	if err != nil {
		return nil, err
	}
	return &model.Link{
		URL: resp.Data.RawURL,
	}, nil
}

func (d *AListV3) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {
	_, err := d.request("/fs/mkdir", http.MethodPost, func(req *resty.Request) {
		req.SetBody(MkdirOrLinkReq{
			Path: path.Join(parentDir.GetPath(), dirName),
		})
	})
	return err
}

func (d *AListV3) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	_, err := d.request("/fs/move", http.MethodPost, func(req *resty.Request) {
		req.SetBody(MoveCopyReq{
			SrcDir: path.Dir(srcObj.GetPath()),
			DstDir: dstDir.GetPath(),
			Names:  []string{srcObj.GetName()},
		})
	})
	return err
}

func (d *AListV3) Rename(ctx context.Context, srcObj model.Obj, newName string) error {
	_, err := d.request("/fs/rename", http.MethodPost, func(req *resty.Request) {
		req.SetBody(RenameReq{
			Path: srcObj.GetPath(),
			Name: newName,
		})
	})
	return err
}

func (d *AListV3) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	_, err := d.request("/fs/copy", http.MethodPost, func(req *resty.Request) {
		req.SetBody(MoveCopyReq{
			SrcDir: path.Dir(srcObj.GetPath()),
			DstDir: dstDir.GetPath(),
			Names:  []string{srcObj.GetName()},
		})
	})
	return err
}

func (d *AListV3) Remove(ctx context.Context, obj model.Obj) error {
	_, err := d.request("/fs/remove", http.MethodPost, func(req *resty.Request) {
		req.SetBody(RemoveReq{
			Dir:   path.Dir(obj.GetPath()),
			Names: []string{obj.GetName()},
		})
	})
	return err
}

func (d *AListV3) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	_, err := d.request("/fs/put", http.MethodPut, func(req *resty.Request) {
		req.SetHeader("File-Path", path.Join(dstDir.GetPath(), stream.GetName())).
			SetHeader("Password", d.MetaPassword).
			SetHeader("Content-Length", strconv.FormatInt(stream.GetSize(), 10)).
			SetContentLength(true).
			SetBody(io.ReadCloser(stream))
	})
	return err
}

//func (d *AList) Other(ctx context.Context, args model.OtherArgs) (interface{}, error) {
//	return nil, errs.NotSupport
//}

var _ driver.Driver = (*AListV3)(nil)
