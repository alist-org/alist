package operations

import (
	"github.com/alist-org/alist/v3/internal/errs"
	stdpath "path"
	"strings"

	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func ActualPath(account driver.Additional, rawPath string) string {
	if i, ok := account.(driver.IRootFolderPath); ok {
		rawPath = stdpath.Join(i.GetRootFolderPath(), rawPath)
	}
	return utils.StandardizePath(rawPath)
}

// GetAccountAndActualPath Get the corresponding account
// for path: remove the virtual path prefix and join the actual root folder if exists
func GetAccountAndActualPath(rawPath string) (driver.Driver, string, error) {
	rawPath = utils.StandardizePath(rawPath)
	if strings.Contains(rawPath, "..") {
		return nil, "", errors.WithStack(errs.RelativePath)
	}
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
