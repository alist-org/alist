package virtual

import (
	"context"
	"io"
	"time"

	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils/random"
)

type Virtual struct {
	model.Storage
	Addition
}

func (d *Virtual) Config() driver.Config {
	return config
}

func (d *Virtual) Init(ctx context.Context) error {
	return nil
}

func (d *Virtual) Drop(ctx context.Context) error {
	return nil
}

func (d *Virtual) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *Virtual) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	var res []model.Obj
	for i := 0; i < d.NumFile; i++ {
		res = append(res, d.genObj(false))
	}
	for i := 0; i < d.NumFolder; i++ {
		res = append(res, d.genObj(true))
	}
	return res, nil
}

type DummyMFile struct {
	io.Reader
}

func (f DummyMFile) Read(p []byte) (n int, err error) {
	return f.Reader.Read(p)
}

func (f DummyMFile) ReadAt(p []byte, off int64) (n int, err error) {
	return f.Reader.Read(p)
}

func (f DummyMFile) Close() error {
	return nil
}

func (DummyMFile) Seek(offset int64, whence int) (int64, error) {
	return offset, nil
}

func (d *Virtual) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	return &model.Link{
		MFile: DummyMFile{Reader: random.Rand},
	}, nil
}

func (d *Virtual) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) (model.Obj, error) {
	dir := &model.Object{
		Name:     dirName,
		Size:     0,
		IsFolder: true,
		Modified: time.Now(),
	}
	return dir, nil
}

func (d *Virtual) Move(ctx context.Context, srcObj, dstDir model.Obj) (model.Obj, error) {
	return srcObj, nil
}

func (d *Virtual) Rename(ctx context.Context, srcObj model.Obj, newName string) (model.Obj, error) {
	obj := &model.Object{
		Name:     newName,
		Size:     srcObj.GetSize(),
		IsFolder: srcObj.IsDir(),
		Modified: time.Now(),
	}
	return obj, nil
}

func (d *Virtual) Copy(ctx context.Context, srcObj, dstDir model.Obj) (model.Obj, error) {
	return srcObj, nil
}

func (d *Virtual) Remove(ctx context.Context, obj model.Obj) error {
	return nil
}

func (d *Virtual) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) (model.Obj, error) {
	file := &model.Object{
		Name:     stream.GetName(),
		Size:     stream.GetSize(),
		Modified: time.Now(),
	}
	return file, nil
}

var _ driver.Driver = (*Virtual)(nil)
