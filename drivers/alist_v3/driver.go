package alist_v3

import (
	"context"
	"encoding/json"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/alist-org/alist/v3/server/handles"
)

type AList_V3 struct {
	model.Storage
	Addition
}

func (d *AList_V3) Config() driver.Config {
	return config
}

func (d *AList_V3) GetAddition() driver.Additional {
	return d.Addition
}

func (d *AList_V3) Init(ctx context.Context, storage model.Storage) error {
	d.Storage = storage
	err := utils.Json.UnmarshalFromString(d.Storage.Addition, &d.Addition)
	if err != nil {
		return err
	}
	// TODO login / refresh token
	//op.MustSaveDriverStorage(d)
	return err
}

func (d *AList_V3) Drop(ctx context.Context) error {
	return nil
}

func (d *AList_V3) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	url := d.Address + "/api/fs/list"
	var resp common.Resp
	_, err := base.RestyClient.R().
		SetResult(&resp).
		SetHeader("Authorization", d.AccessToken).
		SetBody(handles.ListReq{
			PageReq: common.PageReq{
				Page:    1,
				PerPage: 0,
			},
			Path:     dir.GetPath(),
			Password: d.Password,
			Refresh:  false,
		}).Post(url)
	if err != nil {
		return nil, err
	}
	dataStr, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, err
	}
	var data handles.FsListResp
	err = json.Unmarshal(dataStr, &data)
	if err != nil {
		return nil, err
	}
	var files []model.Obj
	for _, f := range data.Content {
		file := model.ObjThumb{
			Object: model.Object{
				Name:     f.Name,
				Modified: f.Modified,
				Size:     f.Size,
				IsFolder: f.IsDir,
			},
		}
		files = append(files, &file)
	}
	return files, nil
}

//func (d *AList) Get(ctx context.Context, path string) (model.Obj, error) {
//	// this is optional
//	return nil, errs.NotImplement
//}

func (d *AList_V3) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	url := d.Address + "/api/fs/get"
	var resp common.Resp
	_, err := base.RestyClient.R().
		SetResult(&resp).
		SetHeader("Authorization", d.AccessToken).
		SetBody(handles.FsGetReq{
			Path:     file.GetPath(),
			Password: d.Password,
		}).Post(url)
	if err != nil {
		return nil, err
	}
	dataStr, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, err
	}
	var data handles.FsGetResp
	err = json.Unmarshal(dataStr, &data)
	if err != nil {
		return nil, err
	}
	return &model.Link{
		URL: data.RawURL,
	}, nil
}

func (d *AList_V3) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {
	return errs.NotImplement
}

func (d *AList_V3) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	return errs.NotImplement
}

func (d *AList_V3) Rename(ctx context.Context, srcObj model.Obj, newName string) error {
	return errs.NotImplement
}

func (d *AList_V3) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	return errs.NotImplement
}

func (d *AList_V3) Remove(ctx context.Context, obj model.Obj) error {
	return errs.NotImplement
}

func (d *AList_V3) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	return errs.NotImplement
}

//func (d *AList) Other(ctx context.Context, args model.OtherArgs) (interface{}, error) {
//	return nil, errs.NotSupport
//}

var _ driver.Driver = (*AList_V3)(nil)
