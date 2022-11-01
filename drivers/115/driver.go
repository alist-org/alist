package _115

import (
	"context"
	"os"

	driver115 "github.com/SheltonZhu/115driver/pkg/driver"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/pkg/errors"
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
	return d.Addition
}

func (d *Pan115) Init(ctx context.Context, storage model.Storage) error {
	d.Storage = storage
	err := utils.Json.UnmarshalFromString(d.Storage.Addition, &d.Addition)
	if err != nil {
		return err
	}
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
	return utils.SliceConvert(files, func(src driver115.File) (model.Obj, error) {
		return src, nil
	})
}

func (d *Pan115) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	downloadInfo, err := d.client.Download(file.(driver115.File).PickCode)
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
	tempFile, err := utils.CreateTempFile(stream.GetReadCloser())
	if err != nil {
		return err
	}
	defer func() {
		_ = tempFile.Close()
		_ = os.Remove(tempFile.Name())
	}()
	return d.client.UploadFastOrByMultipart(dstDir.GetID(), stream.GetName(), stream.GetSize(), tempFile)
}

var _ driver.Driver = (*Pan115)(nil)
