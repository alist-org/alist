package local

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/operations"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/pkg/errors"
)

type Local struct {
	model.Storage
	Addition
}

func (d *Local) Config() driver.Config {
	return config
}

func (d *Local) Init(ctx context.Context, storage model.Storage) error {
	d.Storage = storage
	err := utils.Json.UnmarshalFromString(d.Storage.Addition, &d.Addition)
	if err != nil {
		return errors.Wrap(err, "error while unmarshal addition")
	}
	if !utils.Exists(d.RootFolder) {
		err = errors.Errorf("root folder %s not exists", d.RootFolder)
		d.SetStatus(err.Error())
	} else {
		if !filepath.IsAbs(d.RootFolder) {
			d.RootFolder, err = filepath.Abs(d.RootFolder)
			if err != nil {
				return errors.Wrap(err, "error while get abs path")
			}
		}
		d.SetStatus("OK")
	}
	operations.MustSaveDriverStorage(d)
	return err
}

func (d *Local) Drop(ctx context.Context) error {
	return nil
}

func (d *Local) GetAddition() driver.Additional {
	return d.Addition
}

func (d *Local) List(ctx context.Context, dir model.Obj) ([]model.Obj, error) {
	fullPath := dir.GetID()
	rawFiles, err := ioutil.ReadDir(fullPath)
	if err != nil {
		return nil, errors.Wrapf(err, "error while read dir %s", fullPath)
	}
	var files []model.Obj
	for _, f := range rawFiles {
		if strings.HasPrefix(f.Name(), ".") {
			continue
		}
		file := model.Object{
			Name:     f.Name(),
			Modified: f.ModTime(),
			Size:     f.Size(),
			IsFolder: f.IsDir(),
		}
		files = append(files, &file)
	}
	return files, nil
}

func (d *Local) Get(ctx context.Context, path string) (model.Obj, error) {
	f, err := os.Stat(path)
	if err != nil {
		return nil, errors.Wrapf(err, "error while stat %s", path)
	}
	file := model.Object{
		ID:       path,
		Name:     f.Name(),
		Modified: f.ModTime(),
		Size:     f.Size(),
		IsFolder: f.IsDir(),
	}
	return &file, nil
}

func (d *Local) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	fullPath := file.GetID()
	link := model.Link{
		FilePath: &fullPath,
	}
	return &link, nil
}

func (d *Local) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {
	fullPath := filepath.Join(parentDir.GetID(), dirName)
	err := os.MkdirAll(fullPath, 0700)
	if err != nil {
		return errors.Wrapf(err, "error while make dir %s", fullPath)
	}
	return nil
}

func (d *Local) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	srcPath := srcObj.GetID()
	dstPath := filepath.Join(dstDir.GetID(), srcObj.GetName())
	err := os.Rename(srcPath, dstPath)
	if err != nil {
		return errors.Wrapf(err, "error while move %s to %s", srcPath, dstPath)
	}
	return nil
}

func (d *Local) Rename(ctx context.Context, srcObj model.Obj, newName string) error {
	srcPath := srcObj.GetID()
	dstPath := filepath.Join(filepath.Dir(srcPath), newName)
	err := os.Rename(srcPath, dstPath)
	if err != nil {
		return errors.Wrapf(err, "error while rename %s to %s", srcPath, dstPath)
	}
	return nil
}

func (d *Local) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	srcPath := srcObj.GetID()
	dstPath := filepath.Join(dstDir.GetID(), srcObj.GetName())
	var err error
	if srcObj.IsDir() {
		err = copyDir(srcPath, dstPath)
	} else {
		err = copyFile(srcPath, dstPath)
	}
	if err != nil {
		return errors.Wrapf(err, "error while copy %s to %s", srcPath, dstPath)
	}
	return nil
}

func (d *Local) Remove(ctx context.Context, obj model.Obj) error {
	var err error
	if obj.IsDir() {
		err = os.RemoveAll(obj.GetID())
	} else {
		err = os.Remove(obj.GetID())
	}
	if err != nil {
		return errors.Wrapf(err, "error while remove %s", obj.GetID())
	}
	return nil
}

func (d *Local) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	fullPath := filepath.Join(dstDir.GetID(), stream.GetName())
	out, err := os.Create(fullPath)
	if err != nil {
		return errors.Wrapf(err, "error while create file %s", fullPath)
	}
	defer func() {
		_ = out.Close()
		if errors.Is(err, context.Canceled) {
			_ = os.Remove(fullPath)
		}
	}()
	err = utils.CopyWithCtx(ctx, out, stream)
	if err != nil {
		return errors.Wrapf(err, "error while copy file %s", fullPath)
	}
	return nil
}

func (d *Local) Other(ctx context.Context, args model.OtherArgs) (interface{}, error) {
	return nil, errs.NotSupport
}

var _ driver.Driver = (*Local)(nil)
