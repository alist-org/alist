package store

import (
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/pkg/errors"
)

// CreateAccount just insert account to database
func CreateAccount(account *model.Account) error {
	return errors.WithStack(db.Create(account).Error)
}

// UpdateAccount just update account in database
func UpdateAccount(account *model.Account) error {
	return errors.WithStack(db.Save(account).Error)
}

// DeleteAccountById just delete account from database by id
func DeleteAccountById(id uint) error {
	return errors.WithStack(db.Delete(&model.Account{}, id).Error)
}

// GetAccounts Get all accounts from database
func GetAccounts() ([]model.Account, error) {
	var accounts []model.Account
	if err := db.Order(columnName("index")).Find(&accounts).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return accounts, nil
}

// GetAccountById Get Account by id, used to update account usually
func GetAccountById(id uint) (*model.Account, error) {
	var account model.Account
	account.ID = id
	if err := db.First(&account).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return &account, nil
}
