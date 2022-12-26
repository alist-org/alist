package alist_v3

import (
	"context"
	"io"
	"path"
	"strconv"
	"strings"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/server/common"
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
	// TODO login / refresh token
	//op.MustSaveDriverStorage(d)
	return nil
}

func (d *AListV3) Drop(ctx context.Context) error {
	return nil
}

func (d *AListV3) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	url := d.Address + "/api/fs/list"
	var resp common.Resp[FsListResp]
	_, err := base.RestyClient.R().
		SetResult(&resp).
		SetHeader("Authorization", d.AccessToken).
		SetBody(ListReq{
			PageReq: model.PageReq{
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
	var files []model.Obj
	for _, f := range resp.Data.Content {
		file := model.ObjThumb{
			Object: model.Object{
				Name:     f.Name,
				Modified: f.Modified,
				Size:     f.Size,
				IsFolder: f.IsDir,
			},
			Thumbnail: model.Thumbnail{Thumbnail: f.Thumb},
		}
		files = append(files, &file)
	}
	return files, nil
}

func (d *AListV3) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	url := d.Address + "/api/fs/get"
	var resp common.Resp[FsGetResp]
	_, err := base.RestyClient.R().
		SetResult(&resp).
		SetHeader("Authorization", d.AccessToken).
		SetBody(FsGetReq{
			Path:     file.GetPath(),
			Password: d.Password,
		}).Post(url)
	if err != nil {
		return nil, err
	}
	return &model.Link{
		URL: resp.Data.RawURL,
	}, nil
}

func (d *AListV3) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {
	url := d.Address + "/api/fs/mkdir"
	var resp common.Resp[interface{}]
	_, err := base.RestyClient.R().
		SetResult(&resp).
		SetHeader("Authorization", d.AccessToken).
		SetBody(MkdirOrLinkReq{
			Path: path.Join(parentDir.GetPath(), dirName),
		}).Post(url)
	return checkResp(resp, err)
}

func (d *AListV3) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	url := d.Address + "/api/fs/move"
	var resp common.Resp[interface{}]
	_, err := base.RestyClient.R().
		SetResult(&resp).
		SetHeader("Authorization", d.AccessToken).
		SetBody(MoveCopyReq{
			SrcDir: srcObj.GetPath(),
			DstDir: dstDir.GetPath(),
			Names:  []string{srcObj.GetName()},
		}).Post(url)
	return checkResp(resp, err)
}

func (d *AListV3) Rename(ctx context.Context, srcObj model.Obj, newName string) error {
	url := d.Address + "/api/fs/rename"
	var resp common.Resp[interface{}]
	_, err := base.RestyClient.R().
		SetResult(&resp).
		SetHeader("Authorization", d.AccessToken).
		SetBody(RenameReq{
			Path: srcObj.GetPath(),
			Name: newName,
		}).Post(url)
	return checkResp(resp, err)
}

func (d *AListV3) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	url := d.Address + "/api/fs/copy"
	var resp common.Resp[interface{}]
	_, err := base.RestyClient.R().
		SetResult(&resp).
		SetHeader("Authorization", d.AccessToken).
		SetBody(MoveCopyReq{
			SrcDir: srcObj.GetPath(),
			DstDir: dstDir.GetPath(),
			Names:  []string{srcObj.GetName()},
		}).Post(url)
	return checkResp(resp, err)
}

func (d *AListV3) Remove(ctx context.Context, obj model.Obj) error {
	url := d.Address + "/api/fs/remove"
	var resp common.Resp[interface{}]
	_, err := base.RestyClient.R().
		SetResult(&resp).
		SetHeader("Authorization", d.AccessToken).
		SetBody(RemoveReq{
			Dir:   obj.GetPath(),
			Names: []string{obj.GetName()},
		}).Post(url)
	return checkResp(resp, err)
}

func (d *AListV3) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	url := d.Address + "/api/fs/put"
	var resp common.Resp[interface{}]
	fileBytes, err := io.ReadAll(stream.GetReadCloser())
	if err != nil {
		return nil
	}
	_, err = base.RestyClient.R().SetContext(ctx).
		SetResult(&resp).
		SetHeader("Authorization", d.AccessToken).
		SetHeader("File-Path", path.Join(dstDir.GetPath(), stream.GetName())).
		SetHeader("Password", d.Password).
		SetHeader("Content-Length", strconv.FormatInt(stream.GetSize(), 10)).
		SetBody(fileBytes).Put(url)
	return checkResp(resp, err)
}

//func (d *AList) Other(ctx context.Context, args model.OtherArgs) (interface{}, error) {
//	return nil, errs.NotSupport
//}

var _ driver.Driver = (*AListV3)(nil)
