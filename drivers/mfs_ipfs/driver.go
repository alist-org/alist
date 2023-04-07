package mfs_ipfs

import (
	"context"
	"net/url"

	"github.com/alist-org/alist/v3/drivers/mfs_ipfs/util"
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
)

type MfsIpfs struct {
	model.Storage
	Addition
	mapi *util.MfsAPI
}

func (d *MfsIpfs) Config() driver.Config {
	return config
}

func (d *MfsIpfs) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *MfsIpfs) Init(ctx context.Context) error {
	// TODO login / refresh token
	//op.MustSaveDriverStorage(d)
	util.DefaultPath = conf.Conf.TempDir
	mapi, err := util.NewMfs(&d.CID)
	if err == nil {
		d.mapi = mapi
	}
	return err
}

func (d *MfsIpfs) Drop(ctx context.Context) error {
	if d.mapi == nil {
		return nil
	}
	return d.mapi.Close()
}

func (d *MfsIpfs) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	// TODO return the files list, required
	gateurl, _ := url.Parse(d.Gateway)
	nodelist, err := d.mapi.List(dir.GetPath())
	objlist := []model.Obj{}
	for _, v := range nodelist {
		gateurl.Path = "ipfs/" + v.Id
		gateurl.RawQuery = "filename=" + v.Name
		objlist = append(objlist, &model.ObjectURL{
			Object: model.Object{ID: v.Id, Name: v.Name, Size: v.Size, IsFolder: v.Isdir},
			Url:    model.Url{Url: gateurl.String()},
		})
	}
	return objlist, err
}

func (d *MfsIpfs) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	// TODO return link of file, required
	f, ok := file.(*model.ObjectURL)
	if !ok {
		return nil, errs.NotSupport
	}
	return &model.Link{URL: f.Url.Url}, nil
}

func (d *MfsIpfs) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {
	// TODO create folder, optional
	return errs.NotImplement
}

func (d *MfsIpfs) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	// TODO move obj, optional
	return errs.NotImplement
}

func (d *MfsIpfs) Rename(ctx context.Context, srcObj model.Obj, newName string) error {
	// TODO rename obj, optional
	return errs.NotImplement
}

func (d *MfsIpfs) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	// TODO copy obj, optional
	return errs.NotImplement
}

func (d *MfsIpfs) Remove(ctx context.Context, obj model.Obj) error {
	// TODO remove obj, optional
	return errs.NotImplement
}

func (d *MfsIpfs) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	// TODO upload file, optional
	return errs.NotImplement
}

//func (d *Template) Other(ctx context.Context, args model.OtherArgs) (interface{}, error) {
//	return nil, errs.NotSupport
//}

var _ driver.Driver = (*MfsIpfs)(nil)
