package _115

import (
	"context"
	driver115 "github.com/SheltonZhu/115driver/pkg/driver"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/pkg/errors"
	"os"
)

type Pan115 struct {
	model.Storage
	Addition
	client *driver115.Pan115Client
}

func (d *Pan115) Config() driver.Config {
	return config
}

func (d *Pan115) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *Pan115) Init(ctx context.Context) error {
	return d.login()
}

func (d *Pan115) Drop(ctx context.Context) error {
	return nil
}

func (d *Pan115) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	files, err := d.getFiles(dir.GetID())
	if err != nil && !errors.Is(err, driver115.ErrNotExist) {
		return nil, err
	}
	return utils.SliceConvert(files, func(src FileObj) (model.Obj, error) {
		return &src, nil
	})
}

func (d *Pan115) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	downloadInfo, err := d.client.
		SetUserAgent(driver115.UA115Browser).
		Download(file.(*FileObj).PickCode)
	// recover for upload
	d.client.SetUserAgent(driver115.UA115Desktop)
	if err != nil {
		return nil, err
	}
	link := &model.Link{
		URL:    downloadInfo.Url.Url,
		Header: downloadInfo.Header,
	}
	return link, nil
}

func (d *Pan115) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {
	if _, err := d.client.Mkdir(parentDir.GetID(), dirName); err != nil {
		return err
	}
	return nil
}

func (d *Pan115) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	return d.client.Move(dstDir.GetID(), srcObj.GetID())
}

func (d *Pan115) Rename(ctx context.Context, srcObj model.Obj, newName string) error {
	return d.client.Rename(srcObj.GetID(), newName)
}

func (d *Pan115) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	return d.client.Copy(dstDir.GetID(), srcObj.GetID())
}

func (d *Pan115) Remove(ctx context.Context, obj model.Obj) error {
	return d.client.Delete(obj.GetID())
}

func (d *Pan115) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	tempFile, err := stream.CacheFullInTempFile()
	if err != nil {
		return err
	}
	defer func() {
		_ = tempFile.Close()
	}()
	//TODO: 115 drvier author should update below code, since only few functions is used in the code, no need to use
	// os.File level interface at all
	if result, ok := tempFile.(*os.File); ok {
		return d.client.UploadFastOrByMultipart(dstDir.GetID(), stream.GetName(), stream.GetSize(), result)
	}
	return errs.NotSupport
}

var _ driver.Driver = (*Pan115)(nil)
