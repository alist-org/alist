package local

import (
	"context"

	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/operations"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/pkg/errors"
)

type Driver struct {
	model.Account
	Addition
}

func (d Driver) Config() driver.Config {
	return config
}

func (d *Driver) Init(ctx context.Context, account model.Account) error {
	d.Account = account
	addition := d.Account.Addition
	err := utils.Json.UnmarshalFromString(addition, &d.Addition)
	if err != nil {
		return errors.Wrap(err, "error while unmarshal addition")
	}
	if !utils.Exists(d.RootFolder) {
		err = errors.Errorf("root folder %s not exists", d.RootFolder)
		d.SetStatus(err.Error())
	} else {
		d.SetStatus("OK")
	}
	operations.SaveDriverAccount(d)
	return err
}

func (d *Driver) Drop(ctx context.Context) error {
	return nil
}

func (d *Driver) GetAddition() driver.Additional {
	return d.Addition
}

func (d *Driver) List(ctx context.Context, dir model.Object) ([]model.Object, error) {
	//TODO implement me
	panic("implement me")
}

func (d *Driver) Link(ctx context.Context, file model.Object, args model.LinkArgs) (*model.Link, error) {
	//TODO implement me
	panic("implement me")
}

func (d *Driver) MakeDir(ctx context.Context, parentDir model.Object, dirName string) error {
	//TODO implement me
	panic("implement me")
}

func (d *Driver) Move(ctx context.Context, srcObject, dstDir model.Object) error {
	//TODO implement me
	panic("implement me")
}

func (d *Driver) Rename(ctx context.Context, srcObject model.Object, newName string) error {
	//TODO implement me
	panic("implement me")
}

func (d *Driver) Copy(ctx context.Context, srcObject, dstDir model.Object) error {
	//TODO implement me
	panic("implement me")
}

func (d *Driver) Remove(ctx context.Context, object model.Object) error {
	//TODO implement me
	panic("implement me")
}

func (d *Driver) Put(ctx context.Context, parentDir model.Object, stream model.FileStreamer) error {
	//TODO implement me
	panic("implement me")
}

func (d Driver) Other(ctx context.Context, data interface{}) (interface{}, error) {
	//TODO implement me
	panic("implement me")
}

var _ driver.Driver = (*Driver)(nil)
