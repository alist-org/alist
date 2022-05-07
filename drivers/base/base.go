package base

import "github.com/Xhofe/alist/model"

type Base struct{}

func (b Base) Config() DriverConfig {
	return DriverConfig{}
}

func (b Base) Items() []Item {
	return nil
}

func (b Base) Save(account *model.Account, old *model.Account) error {
	return ErrNotImplement
}

func (b Base) File(path string, account *model.Account) (*model.File, error) {
	return nil, ErrNotImplement
}

func (b Base) Files(path string, account *model.Account) ([]model.File, error) {
	return nil, ErrNotImplement
}

func (b Base) Link(args Args, account *model.Account) (*Link, error) {
	return nil, ErrNotImplement
}

func (b Base) Path(path string, account *model.Account) (*model.File, []model.File, error) {
	return nil, nil, ErrNotImplement
}

func (b Base) Preview(path string, account *model.Account) (interface{}, error) {
	return nil, ErrNotImplement
}

func (b Base) MakeDir(path string, account *model.Account) error {
	return ErrNotImplement
}

func (b Base) Move(src string, dst string, account *model.Account) error {
	return ErrNotImplement
}

func (b Base) Rename(src string, dst string, account *model.Account) error {
	return ErrNotImplement
}

func (b Base) Copy(src string, dst string, account *model.Account) error {
	return ErrNotImplement
}

func (b Base) Delete(path string, account *model.Account) error {
	return ErrNotImplement
}

func (b Base) Upload(file *model.FileStream, account *model.Account) error {
	return ErrNotImplement
}
