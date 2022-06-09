package operations

import (
	"context"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/store"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/pkg/errors"
)

// Although the driver type is stored,
// there is an account in each driver,
// so it should actually be an account, just wrapped by the driver
var accountsMap = map[string]driver.Driver{}

func GetAccountByVirtualPath(virtualPath string) (driver.Driver, error) {
	accountDriver, ok := accountsMap[virtualPath]
	if !ok {
		return nil, errors.Errorf("no virtual path for an account is: %s", virtualPath)
	}
	return accountDriver, nil
}

// CreateAccount Save the account to database so account can get an id
// then instantiate corresponding driver and save it in memory
func CreateAccount(ctx context.Context, account model.Account) error {
	err := store.CreateAccount(&account)
	if err != nil {
		return errors.WithMessage(err, "failed create account in database")
	}
	// already has an id
	driverName := account.Driver
	driverNew, err := GetDriverNew(driverName)
	if err != nil {
		return errors.WithMessage(err, "failed get driver new")
	}
	accountDriver := driverNew()
	err = accountDriver.Init(ctx, account)
	if err != nil {
		return errors.WithMessage(err, "failed init account")
	}
	accountsMap[account.VirtualPath] = accountDriver
	return nil
}

// UpdateAccount update account
// get old account first
// drop the account then reinitialize
func UpdateAccount(ctx context.Context, account model.Account) error {
	oldAccount, err := store.GetAccountById(account.ID)
	if err != nil {
		return errors.WithMessage(err, "failed get old account")
	}
	err = store.UpdateAccount(&account)
	if err != nil {
		return errors.WithMessage(err, "failed update account in database")
	}
	if oldAccount.VirtualPath != account.VirtualPath {
		// virtual path renamed
		delete(accountsMap, oldAccount.VirtualPath)
	}
	accountDriver, err := GetAccountByVirtualPath(oldAccount.VirtualPath)
	if err != nil {
		return errors.WithMessage(err, "failed get account driver")
	}
	err = accountDriver.Drop(ctx)
	if err != nil {
		return errors.WithMessage(err, "failed drop account")
	}
	err = accountDriver.Init(ctx, account)
	if err != nil {
		return errors.WithMessage(err, "failed init account")
	}
	accountsMap[account.VirtualPath] = accountDriver
	return nil
}

// SaveDriverAccount call from specific driver
func SaveDriverAccount(driver driver.Driver) error {
	account := driver.GetAccount()
	addition := driver.GetAddition()
	bytes, err := utils.Json.Marshal(addition)
	if err != nil {
		return errors.Wrap(err, "error while marshal addition")
	}
	account.Addition = string(bytes)
	err = store.UpdateAccount(&account)
	if err != nil {
		return errors.WithMessage(err, "failed update account in database")
	}
	return nil
}
