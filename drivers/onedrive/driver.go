package onedrive

import (
	"context"
	"net/http"
	stdpath "path"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
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
		return errors.Wrap(err, "error while unmarshal addition")
	}
	return d.refreshToken()
}

func (d *Onedrive) Drop(ctx context.Context) error {
	return nil
}

func (d *Onedrive) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	files, err := d.GetFiles(dir.GetID())
	if err != nil {
		return nil, err
	}
	objs := make([]model.Obj, len(files))
	for i := 0; i < len(files); i++ {
		objs[i] = fileToObj(files[i])
	}
	return objs, nil
}

func (d *Onedrive) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	f, err := d.GetFile(file.GetID())
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
	url := d.GetMetaUrl(false, parentDir.GetID()) + "/children"
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
	dst, err := d.GetFile(dstDir.GetID())
	if err != nil {
		return err
	}
	data := base.Json{
		"parentReference": base.Json{
			"id": dst.Id,
		},
		"name": srcObj.GetName(),
	}
	url := d.GetMetaUrl(false, srcObj.GetID())
	_, err = d.Request(url, http.MethodPatch, func(req *resty.Request) {
		req.SetBody(data)
	}, nil)
	return err
}

func (d *Onedrive) Rename(ctx context.Context, srcObj model.Obj, newName string) error {
	dstDir, err := d.GetFile(stdpath.Dir(srcObj.GetID()))
	if err != nil {
		return err
	}
	data := base.Json{
		"parentReference": base.Json{
			"id": dstDir.Id,
		},
		"name": newName,
	}
	url := d.GetMetaUrl(false, srcObj.GetID())
	_, err = d.Request(url, http.MethodPatch, func(req *resty.Request) {
		req.SetBody(data)
	}, nil)
	return err
}

func (d *Onedrive) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	dst, err := d.GetFile(dstDir.GetID())
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
	url := d.GetMetaUrl(false, srcObj.GetID()) + "/copy"
	_, err = d.Request(url, http.MethodPost, func(req *resty.Request) {
		req.SetBody(data)
	}, nil)
	return err
}

func (d *Onedrive) Remove(ctx context.Context, obj model.Obj) error {
	url := d.GetMetaUrl(false, obj.GetID())
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

func (d *Onedrive) Other(ctx context.Context, args model.OtherArgs) (interface{}, error) {
	return nil, errs.NotSupport
}

var _ driver.Driver = (*Onedrive)(nil)
