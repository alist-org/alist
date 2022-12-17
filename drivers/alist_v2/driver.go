package alist_v2

import (
	"context"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/server/common"
)

type AListV2 struct {
	model.Storage
	Addition
}

func (d *AListV2) Config() driver.Config {
	return config
}

func (d *AListV2) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *AListV2) Init(ctx context.Context) error {
	if len(d.Addition.Address) > 0 && string(d.Addition.Address[len(d.Addition.Address)-1]) == "/" {
		d.Addition.Address = d.Addition.Address[0 : len(d.Addition.Address)-1]
	}
	// TODO login / refresh token
	//op.MustSaveDriverStorage(d)
	return nil
}

func (d *AListV2) Drop(ctx context.Context) error {
	return nil
}

func (d *AListV2) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	url := d.Address + "/api/public/path"
	var resp common.Resp[PathResp]
	_, err := base.RestyClient.R().
		SetResult(&resp).
		SetHeader("Authorization", d.AccessToken).
		SetBody(PathReq{
			PageNum:  0,
			PageSize: 0,
			Path:     dir.GetPath(),
			Password: d.Password,
		}).Post(url)
	if err != nil {
		return nil, err
	}
	var files []model.Obj
	for _, f := range resp.Data.Files {
		file := model.ObjThumb{
			Object: model.Object{
				Name:     f.Name,
				Modified: *f.UpdatedAt,
				Size:     f.Size,
				IsFolder: f.Type == 1,
			},
			Thumbnail: model.Thumbnail{Thumbnail: f.Thumbnail},
		}
		files = append(files, &file)
	}
	return files, nil
}

func (d *AListV2) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	url := d.Address + "/api/public/path"
	var resp common.Resp[PathResp]
	_, err := base.RestyClient.R().
		SetResult(&resp).
		SetHeader("Authorization", d.AccessToken).
		SetBody(PathReq{
			PageNum:  0,
			PageSize: 0,
			Path:     file.GetPath(),
			Password: d.Password,
		}).Post(url)
	if err != nil {
		return nil, err
	}
	return &model.Link{
		URL: resp.Data.Files[0].Url,
	}, nil
}

func (d *AListV2) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {
	return errs.NotImplement
}

func (d *AListV2) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	return errs.NotImplement
}

func (d *AListV2) Rename(ctx context.Context, srcObj model.Obj, newName string) error {
	return errs.NotImplement
}

func (d *AListV2) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	return errs.NotImplement
}

func (d *AListV2) Remove(ctx context.Context, obj model.Obj) error {
	return errs.NotImplement
}

func (d *AListV2) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	return errs.NotImplement
}

//func (d *AList) Other(ctx context.Context, args model.OtherArgs) (interface{}, error) {
//	return nil, errs.NotSupport
//}

var _ driver.Driver = (*AListV2)(nil)
