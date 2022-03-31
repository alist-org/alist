package common

import (
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/drivers/operate"
	"github.com/Xhofe/alist/model"
	log "github.com/sirupsen/logrus"
)

func Path(rawPath string) (*model.File, []model.File, *model.Account, base.Driver, string, error) {
	account, path, driver, err := ParsePath(rawPath)
	if err != nil {
		if err.Error() == "path not found" {
			accountFiles := model.GetAccountFilesByPath(rawPath)
			if len(accountFiles) != 0 {
				return nil, accountFiles, nil, nil, path, nil
			}
		}
		return nil, nil, nil, nil, "", err
	}
	log.Debugln("use account: ", account.Name)
	file, files, err := operate.Path(driver, account, path)
	if err != nil {
		return nil, nil, nil, nil, "", err
	}
	if file != nil {
		return file, nil, account, driver, path, nil
	} else {
		accountFiles := model.GetAccountFilesByPath(rawPath)
		files = append(files, accountFiles...)
		return nil, files, account, driver, path, nil
	}
}
