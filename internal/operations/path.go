package operations

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"path"
	"strings"
)

func ActualPath(account driver.Additional, rawPath string) string {
	if i, ok := account.(driver.IRootFolderPath); ok {
		rawPath = path.Join(i.GetRootFolder(), rawPath)
	}
	return utils.StandardizationPath(rawPath)
}

// GetAccountAndActualPath Get the corresponding account, and remove the virtual path prefix in path
func GetAccountAndActualPath(rawPath string) (driver.Driver, string, error) {
	rawPath = utils.StandardizationPath(rawPath)
	account := GetBalancedAccount(rawPath)
	if account == nil {
		return nil, "", errors.Errorf("can't find account with rawPath: %s", rawPath)
	}
	log.Debugln("use account: ", account.GetAccount().VirtualPath)
	virtualPath := utils.GetActualVirtualPath(account.GetAccount().VirtualPath)
	actualPath := strings.TrimPrefix(rawPath, virtualPath)
	actualPath = ActualPath(account.GetAddition(), actualPath)
	return account, actualPath, nil
}
