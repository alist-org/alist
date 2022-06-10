package operations

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"strings"
)

// GetAccountAndActualPath Get the corresponding account, and remove the virtual path prefix in path
func GetAccountAndActualPath(path string) (driver.Driver, string, error) {
	path = utils.StandardizationPath(path)
	account := GetBalancedAccount(path)
	if account == nil {
		return nil, "", errors.Errorf("can't find account with path: %s", path)
	}
	log.Debugln("use account: ", account.GetAccount().VirtualPath)
	virtualPath := utils.GetActualVirtualPath(account.GetAccount().VirtualPath)
	actualPath := utils.StandardizationPath(strings.TrimPrefix(path, virtualPath))
	return account, actualPath, nil
}
