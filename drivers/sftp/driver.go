package sftp

import (
	"context"
	"os"
	"path"

	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/pkg/sftp"
)

type SFTP struct {
	model.Storage
	Addition
	client *sftp.Client
}

func (d *SFTP) Config() driver.Config {
	return config
}

func (d *SFTP) GetAddition() driver.Additional {
	return d.Addition
}

func (d *SFTP) Init(ctx context.Context, storage model.Storage) error {
	d.Storage = storage
	err := utils.Json.UnmarshalFromString(d.Storage.Addition, &d.Addition)
	if err != nil {
		return err
	}
	return d.initClient()
}

func (d *SFTP) Drop(ctx context.Context) error {
	if d.client != nil {
		_ = d.client.Close()
	}
	return nil
}

func (d *SFTP) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	files, err := d.client.ReadDir(dir.GetPath())
	if err != nil {
		return nil, err
	}
	return utils.SliceConvert(files, func(src os.FileInfo) (model.Obj, error) {
		return fileToObj(src), nil
	})
}

//func (d *SFTP) Get(ctx context.Context, path string) (model.Obj, error) {
//	// this is optional
//	return nil, errs.NotImplement
//}

func (d *SFTP) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	remoteFile, err := d.client.Open(file.GetPath())
	if err != nil {
		return nil, err
	}
	return &model.Link{
		Data: remoteFile,
	}, nil
}

func (d *SFTP) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {
	return d.client.MkdirAll(path.Join(parentDir.GetPath(), dirName))
}

func (d *SFTP) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	return d.client.Rename(srcObj.GetPath(), path.Join(dstDir.GetPath(), srcObj.GetName()))
}

func (d *SFTP) Rename(ctx context.Context, srcObj model.Obj, newName string) error {
	return d.client.Rename(srcObj.GetPath(), path.Join(path.Dir(srcObj.GetPath()), newName))
}

func (d *SFTP) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	return errs.NotSupport
}

func (d *SFTP) Remove(ctx context.Context, obj model.Obj) error {
	return d.remove(obj.GetPath())
}

func (d *SFTP) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	dstFile, err := d.client.Create(path.Join(dstDir.GetPath(), stream.GetName()))
	if err != nil {
		return err
	}
	defer func() {
		_ = dstFile.Close()
	}()
	err = utils.CopyWithCtx(ctx, dstFile, stream, stream.GetSize(), up)
	return err
}

var _ driver.Driver = (*SFTP)(nil)
