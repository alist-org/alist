package operations

import (
	"context"
	log "github.com/sirupsen/logrus"
	"sort"
	"strings"
	"time"

	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/generic_sync"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/pkg/errors"
)

// Although the driver type is stored,
// there is an account in each driver,
// so it should actually be an account, just wrapped by the driver
var accountsMap generic_sync.MapOf[string, driver.Driver]

func GetAccountByVirtualPath(virtualPath string) (driver.Driver, error) {
	accountDriver, ok := accountsMap.Load(virtualPath)
	if !ok {
		return nil, errors.Errorf("no virtual path for an account is: %s", virtualPath)
	}
	return accountDriver, nil
}

// CreateAccount Save the account to database so account can get an id
// then instantiate corresponding driver and save it in memory
func CreateAccount(ctx context.Context, account model.Account) error {
	account.Modified = time.Now()
	account.VirtualPath = utils.StandardizePath(account.VirtualPath)
	var err error
	// check driver first
	driverName := account.Driver
	driverNew, err := GetDriverNew(driverName)
	if err != nil {
		return errors.WithMessage(err, "failed get driver new")
	}
	accountDriver := driverNew()
	// insert account to database
	err = db.CreateAccount(&account)
	if err != nil {
		return errors.WithMessage(err, "failed create account in database")
	}
	// already has an id
	err = accountDriver.Init(ctx, account)
	if err != nil {
		return errors.WithMessage(err, "failed init account but account is already created")
	}
	log.Debugf("account %+v is created", accountDriver)
	accountsMap.Store(account.VirtualPath, accountDriver)
	return nil
}

// UpdateAccount update account
// get old account first
// drop the account then reinitialize
func UpdateAccount(ctx context.Context, account model.Account) error {
	oldAccount, err := db.GetAccountById(account.ID)
	if err != nil {
		return errors.WithMessage(err, "failed get old account")
	}
	if oldAccount.Driver != account.Driver {
		return errors.Errorf("driver cannot be changed")
	}
	account.Modified = time.Now()
	account.VirtualPath = utils.StandardizePath(account.VirtualPath)
	err = db.UpdateAccount(&account)
	if err != nil {
		return errors.WithMessage(err, "failed update account in database")
	}
	accountDriver, err := GetAccountByVirtualPath(oldAccount.VirtualPath)
	if oldAccount.VirtualPath != account.VirtualPath {
		// virtual path renamed, need to drop the account
		accountsMap.Delete(oldAccount.VirtualPath)
	}
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
	accountsMap.Store(account.VirtualPath, accountDriver)
	return nil
}

func DeleteAccountById(ctx context.Context, id uint) error {
	account, err := db.GetAccountById(id)
	if err != nil {
		return errors.WithMessage(err, "failed get account")
	}
	accountDriver, err := GetAccountByVirtualPath(account.VirtualPath)
	if err != nil {
		return errors.WithMessage(err, "failed get account driver")
	}
	// drop the account in the driver
	if err := accountDriver.Drop(ctx); err != nil {
		return errors.WithMessage(err, "failed drop account")
	}
	// delete the account in the database
	if err := db.DeleteAccountById(id); err != nil {
		return errors.WithMessage(err, "failed delete account in database")
	}
	// delete the account in the memory
	accountsMap.Delete(account.VirtualPath)
	return nil
}

// MustSaveDriverAccount call from specific driver
func MustSaveDriverAccount(driver driver.Driver) {
	err := saveDriverAccount(driver)
	if err != nil {
		log.Errorf("failed save driver account: %s", err)
	}
}

func saveDriverAccount(driver driver.Driver) error {
	account := driver.GetAccount()
	addition := driver.GetAddition()
	bytes, err := utils.Json.Marshal(addition)
	if err != nil {
		return errors.Wrap(err, "error while marshal addition")
	}
	account.Addition = string(bytes)
	err = db.UpdateAccount(&account)
	if err != nil {
		return errors.WithMessage(err, "failed update account in database")
	}
	return nil
}

// GetAccountsByPath get account by longest match path, contains balance account.
// for example, there is /a/b,/a/c,/a/d/e,/a/d/e.balance
// GetAccountsByPath(/a/d/e/f) => /a/d/e,/a/d/e.balance
func getAccountsByPath(path string) []driver.Driver {
	accounts := make([]driver.Driver, 0)
	curSlashCount := 0
	accountsMap.Range(func(key string, value driver.Driver) bool {
		virtualPath := utils.GetActualVirtualPath(value.GetAccount().VirtualPath)
		if virtualPath == "/" {
			virtualPath = ""
		}
		// not this
		if path != virtualPath && !strings.HasPrefix(path, virtualPath+"/") {
			return true
		}
		slashCount := strings.Count(virtualPath, "/")
		// not the longest match
		if slashCount < curSlashCount {
			return true
		}
		if slashCount > curSlashCount {
			accounts = accounts[:0]
			curSlashCount = slashCount
		}
		accounts = append(accounts, value)
		return true
	})
	// make sure the order is the same for same input
	sort.Slice(accounts, func(i, j int) bool {
		return accounts[i].GetAccount().VirtualPath < accounts[j].GetAccount().VirtualPath
	})
	return accounts
}

// GetAccountVirtualFilesByPath Obtain the virtual file generated by the account according to the path
// for example, there are: /a/b,/a/c,/a/d/e,/a/b.balance1,/av
// GetAccountVirtualFilesByPath(/a) => b,c,d
func GetAccountVirtualFilesByPath(prefix string) []model.Obj {
	files := make([]model.Obj, 0)
	accounts := accountsMap.Values()
	sort.Slice(accounts, func(i, j int) bool {
		if accounts[i].GetAccount().Index == accounts[j].GetAccount().Index {
			return accounts[i].GetAccount().VirtualPath < accounts[j].GetAccount().VirtualPath
		}
		return accounts[i].GetAccount().Index < accounts[j].GetAccount().Index
	})
	prefix = utils.StandardizePath(prefix)
	set := make(map[string]interface{})
	for _, v := range accounts {
		// TODO should save a balanced account
		// balance account
		if utils.IsBalance(v.GetAccount().VirtualPath) {
			continue
		}
		virtualPath := v.GetAccount().VirtualPath
		if len(virtualPath) <= len(prefix) {
			continue
		}
		// not prefixed with `prefix`
		if !strings.HasPrefix(virtualPath, prefix+"/") && prefix != "/" {
			continue
		}
		name := strings.Split(strings.TrimPrefix(virtualPath, prefix), "/")[1]
		if _, ok := set[name]; ok {
			continue
		}
		files = append(files, model.Object{
			Name:     name,
			Size:     0,
			Modified: v.GetAccount().Modified,
		})
		set[name] = nil
	}
	return files
}

var balanceMap generic_sync.MapOf[string, int]

// GetBalancedAccount get account by path
func GetBalancedAccount(path string) driver.Driver {
	path = utils.StandardizePath(path)
	accounts := getAccountsByPath(path)
	accountNum := len(accounts)
	switch accountNum {
	case 0:
		return nil
	case 1:
		return accounts[0]
	default:
		virtualPath := utils.GetActualVirtualPath(accounts[0].GetAccount().VirtualPath)
		cur, ok := balanceMap.Load(virtualPath)
		i := 0
		if ok {
			i = cur
			i = (i + 1) % accountNum
			balanceMap.Store(virtualPath, i)
		} else {
			balanceMap.Store(virtualPath, i)
		}
		return accounts[i]
	}
}
