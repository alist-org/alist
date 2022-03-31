package common

import (
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/drivers/operate"
	"github.com/Xhofe/alist/model"
)

func Path(rawPath string) (*model.File, []model.File, *model.Account, base.Driver, string, error) {
	account, path, driver, err := ParsePath(rawPath)
	if err != nil {
		if err.Error() == "path not found" {
			accountFiles, err := model.GetAccountFilesByPath(rawPath)
			if err != nil {
				return nil, nil, nil, nil, "", err
			}
			if len(accountFiles) != 0 {
				return nil, accountFiles, nil, nil, path, nil
			}
		}
		return nil, nil, nil, nil, "", err
	}
	file, files, err := operate.Path(driver, account, path)
	if err != nil {
		return nil, nil, nil, nil, "", err
	}
	if file != nil {
		return file, nil, account, driver, path, nil
	} else {
		accountFiles, err := model.GetAccountFilesByPath(rawPath)
		if err != nil {
			return nil, nil, nil, nil, "", err
		}
		files = append(files, accountFiles...)
		return nil, files, account, driver, path, nil
	}
}
