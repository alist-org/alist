package local

import (
	"context"
	"errors"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
)

type Driver struct {
	model.Account
	Addition
}

func (d Driver) Config() driver.Config {
	return config
}

func (d *Driver) Init(ctx context.Context, account model.Account) error {
	addition := d.Account.Addition
	err := utils.Json.UnmarshalFromString(addition, d.Addition)
	if err != nil {
		return errors.New("error")
	}
	return nil
}

func (d *Driver) Update(ctx context.Context, account model.Account) error {
	//TODO implement me
	panic("implement me")
}

func (d *Driver) Drop(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (d *Driver) GetAccount() model.Account {
	//TODO implement me
	panic("implement me")
}

func (d *Driver) File(ctx context.Context, path string) (*driver.FileInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (d *Driver) List(ctx context.Context, path string) ([]driver.FileInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (d *Driver) Link(ctx context.Context, args driver.LinkArgs) (*driver.Link, error) {
	//TODO implement me
	panic("implement me")
}

func (d *Driver) MakeDir(ctx context.Context, path string) error {
	//TODO implement me
	panic("implement me")
}

func (d *Driver) Move(ctx context.Context, src, dst string) error {
	//TODO implement me
	panic("implement me")
}

func (d *Driver) Rename(ctx context.Context, src, dst string) error {
	//TODO implement me
	panic("implement me")
}

func (d *Driver) Copy(ctx context.Context, src, dst string) error {
	//TODO implement me
	panic("implement me")
}

func (d *Driver) Remove(ctx context.Context, path string) error {
	//TODO implement me
	panic("implement me")
}

func (d *Driver) Put(ctx context.Context, stream driver.FileStream, parentPath string) error {
	//TODO implement me
	panic("implement me")
}

var _ driver.Driver = (*Driver)(nil)

func init() {
	driver.RegisterDriver(config.Name, New)
}
