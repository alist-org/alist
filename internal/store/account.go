package store

import (
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/pkg/errors"
)

// why don't need `cache` for account?
// because all account store in `operations.accountsMap`
// the most of the read operation is from `operations.accountsMap`
// just for persistence in database

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

// GetAccounts Get all accounts from database order by index
func GetAccounts(pageIndex, pageSize int) ([]model.Account, int64, error) {
	accountDB := db.Model(&model.Account{})
	var count int64
	if err := accountDB.Count(&count).Error; err != nil {
		return nil, 0, errors.Wrapf(err, "failed get accounts count")
	}
	var accounts []model.Account
	if err := accountDB.Order(columnName("index")).Offset((pageIndex - 1) * pageSize).Limit(pageSize).Find(&accounts).Error; err != nil {
		return nil, 0, errors.WithStack(err)
	}
	return accounts, count, nil
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
