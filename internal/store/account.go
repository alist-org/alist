package store

import "github.com/alist-org/alist/v3/internal/model"

type Account interface {
	Create(account model.Account) error
	Update(account model.Account) error
	Delete(id uint) error
	GetByID(id uint) (*model.Account, error)
	List() ([]model.Account, error)
}
