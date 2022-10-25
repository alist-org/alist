package smb

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"

	"github.com/hirochachacha/go-smb2"
)

type SMB struct {
	model.Storage
	Addition
	fs *smb2.Share
}

func (d *SMB) Config() driver.Config {
	return config
}

func (d *SMB) GetAddition() driver.Additional {
	return d.Addition
}

func (d *SMB) Init(ctx context.Context, storage model.Storage) error {
	d.Storage = storage
	err := utils.Json.UnmarshalFromString(d.Storage.Addition, &d.Addition)
	if err != nil {
		return err
	}
	return d.initFS()
}

func (d *SMB) Drop(ctx context.Context) error {
	if d.fs != nil {
		_ = d.fs.Umount()
	}
	return nil
}

func (d *SMB) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	fullPath := d.getSMBPath(dir)
	rawFiles, err := d.fs.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}
	var files []model.Obj
	for _, f := range rawFiles {
		file := model.ObjThumb{
			Object: model.Object{
				Name:     f.Name(),
				Modified: f.ModTime(),
				Size:     f.Size(),
				IsFolder: f.IsDir(),
			},
		}
		files = append(files, &file)
	}
	return files, nil
}

//func (d *SMB) Get(ctx context.Context, path string) (model.Obj, error) {
//	// this is optional
//	return nil, errs.NotImplement
//}

func (d *SMB) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	fullPath := d.getSMBPath(file)
	remoteFile, err := d.fs.Open(fullPath)
	if err != nil {
		return nil, err
	}
	return &model.Link{
		Data: remoteFile,
	}, nil
}

func (d *SMB) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {
	fullPath := filepath.Join(d.getSMBPath(parentDir), dirName)
	err := d.fs.MkdirAll(fullPath, 0700)
	if err != nil {
		return err
	}
	return nil
}

func (d *SMB) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	srcPath := d.getSMBPath(srcObj)
	dstPath := filepath.Join(d.getSMBPath(dstDir), srcObj.GetName())
	err := d.fs.Rename(srcPath, dstPath)
	if err != nil {
		return err
	}
	return nil
}

func (d *SMB) Rename(ctx context.Context, srcObj model.Obj, newName string) error {
	srcPath := d.getSMBPath(srcObj)
	dstPath := filepath.Join(filepath.Dir(srcPath), newName)
	err := d.fs.Rename(srcPath, dstPath)
	if err != nil {
		return err
	}
	return nil
}

func (d *SMB) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	srcPath := d.getSMBPath(srcObj)
	dstPath := filepath.Join(d.getSMBPath(dstDir), srcObj.GetName())
	var err error
	if srcObj.IsDir() {
		err = d.CopyDir(srcPath, dstPath)
	} else {
		err = d.CopyFile(srcPath, dstPath)
	}
	if err != nil {
		return err
	}
	return nil
}

func (d *SMB) Remove(ctx context.Context, obj model.Obj) error {
	var err error
	fullPath := d.getSMBPath(obj)
	if obj.IsDir() {
		err = d.fs.RemoveAll(fullPath)
	} else {
		err = d.fs.Remove(fullPath)
	}
	if err != nil {
		return err
	}
	return nil
}

func (d *SMB) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	fullPath := filepath.Join(d.getSMBPath(dstDir), stream.GetName())
	out, err := d.fs.Create(fullPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = out.Close()
		if errors.Is(err, context.Canceled) {
			_ = d.fs.Remove(fullPath)
		}
	}()
	err = utils.CopyWithCtx(ctx, out, stream, stream.GetSize(), up)
	if err != nil {
		return err
	}
	return nil
}

//func (d *SMB) Other(ctx context.Context, args model.OtherArgs) (interface{}, error) {
//	return nil, errs.NotSupport
//}

var _ driver.Driver = (*SMB)(nil)
