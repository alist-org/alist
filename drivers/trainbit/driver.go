package trainbit

import (
	"context"
	"net/http"

	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
)

type Trainbit struct {
	model.Storage
	Addition
}

func (d *Trainbit) Config() driver.Config {
	return config
}

func (d *Trainbit) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *Trainbit) Init(ctx context.Context) error {
	//op.MustSaveDriverStorage(d)
	http.DefaultClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
        return http.ErrUseLastResponse
    }
	return nil
}

func (d *Trainbit) Drop(ctx context.Context) error {
	return nil
}

func (d *Trainbit) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	objectList, err := readFolder(dir.GetID(), d.AUSHELLPORTAL, d.ApiKey)
	if err != nil {
		return nil, err
	}
	return objectList, nil
}

func (d *Trainbit) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	link, err := getDownloadLink(file.GetID(), d.AUSHELLPORTAL)
	if err != nil {
		return nil, err
	}
	return &model.Link{
		URL: link,
	}, nil
}

func (d *Trainbit) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {
	// TODO create folder
	return errs.NotImplement
}

func (d *Trainbit) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	// TODO move obj
	return errs.NotImplement
}

func (d *Trainbit) Rename(ctx context.Context, srcObj model.Obj, newName string) error {
	// TODO rename obj
	return errs.NotImplement
}

func (d *Trainbit) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	// TODO copy obj
	return errs.NotImplement
}

func (d *Trainbit) Remove(ctx context.Context, obj model.Obj) error {
	// TODO remove obj
	return errs.NotImplement
}

func (d *Trainbit) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	// TODO upload file
	return errs.NotImplement
}

//func (d *Trainbit) Other(ctx context.Context, args model.OtherArgs) (interface{}, error) {
//	return nil, errs.NotSupport
//}

var _ driver.Driver = (*Trainbit)(nil)
