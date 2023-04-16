package mfs_ipfs

import (
	"context"
	"log"
	"net/url"
	"path"

	"github.com/alist-org/alist/v3/drivers/mfs_ipfs/util"
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/sirupsen/logrus"
)

type MfsIpfs struct {
	model.Storage
	Addition
	mapi    *util.MfsAPI
	gateurl *url.URL
}
type refid struct {
	d  *MfsIpfs
	id *string
}

func (r *refid) Get() (id string) { return *r.id }
func (r *refid) Set(id string) {
	*r.id = id
	op.MustSaveDriverStorage(r.d)
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
	var err error
	d.gateurl, _ = url.Parse(d.Gateway)
	d.mapi, err = util.NewMfs(
		d.Endpoint, d.JWToken,
		&refid{d: d, id: &d.Addition.CID}, &refid{d: d, id: &d.Addition.PinID},
		log.New(logrus.StandardLogger().WriterLevel(logrus.WarnLevel), "IPFS", log.LstdFlags),
	)
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
	nodelist, err := d.mapi.List(dir.GetPath())
	objlist := []model.Obj{}
	for _, v := range nodelist {
		gateurl := *d.gateurl
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
	link, ok := model.GetUrl(file)
	if !ok {
		return nil, errs.NotSupport
	}
	return &model.Link{URL: link}, nil
}

func (d *MfsIpfs) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {
	// TODO create folder, optional
	return d.mapi.Mkdir(path.Join(parentDir.GetPath(), dirName))
}

func (d *MfsIpfs) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	// TODO move obj, optional
	return d.mapi.Mv(srcObj.GetPath(), path.Join(dstDir.GetPath(), srcObj.GetName()))
}

func (d *MfsIpfs) Rename(ctx context.Context, srcObj model.Obj, newName string) error {
	// TODO rename obj, optional
	return d.mapi.Mv(srcObj.GetPath(), path.Join(path.Dir(srcObj.GetPath()), newName))
}

func (d *MfsIpfs) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	// TODO copy obj, optional
	return d.mapi.Put(path.Join(dstDir.GetPath(), srcObj.GetName()), srcObj.GetID(), nil)
}

func (d *MfsIpfs) Remove(ctx context.Context, obj model.Obj) error {
	// TODO remove obj, optional
	return d.mapi.Unlink(path.Dir(obj.GetPath()), obj.GetName())
}

func (d *MfsIpfs) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	// TODO upload file, optional
	return d.mapi.Put(path.Join(dstDir.GetPath(), stream.GetName()), stream.GetID(), stream.GetReadCloser())
}

//func (d *Template) Other(ctx context.Context, args model.OtherArgs) (interface{}, error) {
//	return nil, errs.NotSupport
//}

var _ driver.Driver = (*MfsIpfs)(nil)
