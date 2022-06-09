package operations

import (
	"context"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/store"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/pkg/errors"
	"sort"
	"strings"
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

var balance = ".balance"

// GetAccountsByPath get account by longest match path, contains balance account.
// for example, there is /a/b,/a/c,/a/d/e,/a/d/e.balance
// GetAccountsByPath(/a/d/e/f) => /a/d/e,/a/d/e.balance
func GetAccountsByPath(path string) []driver.Driver {
	accounts := make([]driver.Driver, 0)
	curSlashCount := 0
	for _, v := range accountsMap {
		virtualPath := utils.StandardizationPath(v.GetAccount().VirtualPath)
		bIndex := strings.LastIndex(virtualPath, balance)
		if bIndex != -1 {
			virtualPath = virtualPath[:bIndex]
		}
		if virtualPath == "/" {
			virtualPath = ""
		}
		// not this
		if path != virtualPath && !strings.HasPrefix(path, virtualPath+"/") {
			continue
		}
		slashCount := strings.Count(virtualPath, "/")
		// not the longest match
		if slashCount < curSlashCount {
			continue
		}
		if slashCount > curSlashCount {
			accounts = accounts[:0]
			curSlashCount = slashCount
		}
		accounts = append(accounts, v)
	}
	// make sure the order is the same for same input
	sort.Slice(accounts, func(i, j int) bool {
		return accounts[i].GetAccount().VirtualPath < accounts[j].GetAccount().VirtualPath
	})
	return accounts
}
