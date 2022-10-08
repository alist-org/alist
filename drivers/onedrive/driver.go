package onedrive

import (
	"context"
	"net/http"
	stdpath "path"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/go-resty/resty/v2"
)

type Onedrive struct {
	model.Storage
	Addition
	AccessToken string
}

func (d *Onedrive) Config() driver.Config {
	return config
}

func (d *Onedrive) GetAddition() driver.Additional {
	return d.Addition
}

func (d *Onedrive) Init(ctx context.Context, storage model.Storage) error {
	d.Storage = storage
	err := utils.Json.UnmarshalFromString(d.Storage.Addition, &d.Addition)
	if err != nil {
		return err
	}
	if d.ChunkSize < 1 {
		d.ChunkSize = 5
	}
	return d.refreshToken()
}

func (d *Onedrive) Drop(ctx context.Context) error {
	return nil
}

func (d *Onedrive) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	files, err := d.getFiles(dir.GetPath())
	if err != nil {
		return nil, err
	}
	return utils.SliceConvert(files, func(src File) (model.Obj, error) {
		return fileToObj(src), nil
	})
}

func (d *Onedrive) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	f, err := d.GetFile(file.GetPath())
	if err != nil {
		return nil, err
	}
	if f.File == nil {
		return nil, errs.NotFile
	}
	return &model.Link{
		URL: f.Url,
	}, nil
}

func (d *Onedrive) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {
	url := d.GetMetaUrl(false, parentDir.GetPath()) + "/children"
	data := base.Json{
		"name":                              dirName,
		"folder":                            base.Json{},
		"@microsoft.graph.conflictBehavior": "rename",
	}
	_, err := d.Request(url, http.MethodPost, func(req *resty.Request) {
		req.SetBody(data)
	}, nil)
	return err
}

func (d *Onedrive) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	data := base.Json{
		"parentReference": base.Json{
			"id": dstDir.GetID(),
		},
		"name": srcObj.GetName(),
	}
	url := d.GetMetaUrl(false, srcObj.GetPath())
	_, err := d.Request(url, http.MethodPatch, func(req *resty.Request) {
		req.SetBody(data)
	}, nil)
	return err
}

func (d *Onedrive) Rename(ctx context.Context, srcObj model.Obj, newName string) error {
	dstDir, err := op.Get(ctx, d, stdpath.Dir(srcObj.GetPath()))
	if err != nil {
		return err
	}
	data := base.Json{
		"parentReference": base.Json{
			"id": dstDir.GetID(),
		},
		"name": newName,
	}
	url := d.GetMetaUrl(false, srcObj.GetPath())
	_, err = d.Request(url, http.MethodPatch, func(req *resty.Request) {
		req.SetBody(data)
	}, nil)
	return err
}

func (d *Onedrive) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	dst, err := d.GetFile(dstDir.GetPath())
	if err != nil {
		return err
	}
	data := base.Json{
		"parentReference": base.Json{
			"driveId": dst.ParentReference.DriveId,
			"id":      dst.Id,
		},
		"name": srcObj.GetName(),
	}
	url := d.GetMetaUrl(false, srcObj.GetPath()) + "/copy"
	_, err = d.Request(url, http.MethodPost, func(req *resty.Request) {
		req.SetBody(data)
	}, nil)
	return err
}

func (d *Onedrive) Remove(ctx context.Context, obj model.Obj) error {
	url := d.GetMetaUrl(false, obj.GetPath())
	_, err := d.Request(url, http.MethodDelete, nil, nil)
	return err
}

func (d *Onedrive) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	var err error
	if stream.GetSize() <= 4*1024*1024 {
		err = d.upSmall(dstDir, stream)
	} else {
		err = d.upBig(ctx, dstDir, stream, up)
	}
	return err
}

var _ driver.Driver = (*Onedrive)(nil)
