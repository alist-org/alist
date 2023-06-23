package template

import (
	"context"

	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
)

type Dropbox struct {
	model.Storage
	Addition

	dbx files.Client
}

func (d *Dropbox) Config() driver.Config {
	return config
}

func (d *Dropbox) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *Dropbox) Init(ctx context.Context) error {
	cfg := dropbox.Config{
		Token: d.AccessToken,
	}
	d.dbx = files.New(cfg)
	return nil
}

func (d *Dropbox) Drop(ctx context.Context) error {
	return nil
}

func (d *Dropbox) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	// TODO return the files list, required
	return nil, errs.NotImplement
}

func (d *Dropbox) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	// TODO return link of file, required
	return nil, errs.NotImplement
}

func (d *Dropbox) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {
	// TODO create folder, optional
	return errs.NotImplement
}

func (d *Dropbox) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	// TODO move obj, optional
	return errs.NotImplement
}

func (d *Dropbox) Rename(ctx context.Context, srcObj model.Obj, newName string) error {
	// TODO rename obj, optional
	return errs.NotImplement
}

func (d *Dropbox) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	// TODO copy obj, optional
	return errs.NotImplement
}

func (d *Dropbox) Remove(ctx context.Context, obj model.Obj) error {
	// TODO remove obj, optional
	return errs.NotImplement
}

func (d *Dropbox) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	// TODO upload file, optional
	return errs.NotImplement
}

//func (d *Dropbox) Other(ctx context.Context, args model.OtherArgs) (interface{}, error) {
//	return nil, errs.NotSupport
//}

var _ driver.Driver = (*Dropbox)(nil)
