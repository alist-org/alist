package local

import (
	"context"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
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
	err := utils.Json.UnmarshalFromString(addition, d.Addition)
	if err != nil {
		return errors.Wrap(err, "error while unmarshal addition")
	}
	return nil
}

func (d *Driver) Drop(ctx context.Context) error {
	return nil
}

func (d *Driver) GetAddition() driver.Additional {
	return d.Addition
}

func (d *Driver) File(ctx context.Context, path string) (driver.FileInfo, error) {
	fullPath := filepath.Join(d.RootFolder, path)
	if !utils.Exists(fullPath) {
		return nil, errors.WithStack(driver.ErrorObjectNotFound)
	}
	f, err := os.Stat(fullPath)
	if err != nil {
		return nil, err
	}
	return model.File{
		Name:     f.Name(),
		Size:     uint64(f.Size()),
		Modified: f.ModTime(),
		IsFolder: f.IsDir(),
	}, nil
}

func (d *Driver) List(ctx context.Context, path string) ([]driver.FileInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (d *Driver) Link(ctx context.Context, args driver.LinkArgs) (*driver.Link, error) {
	//TODO implement me
	panic("implement me")
}

func (d Driver) Other(ctx context.Context, data interface{}) (interface{}, error) {
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
