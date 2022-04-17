package webdav

import "github.com/Xhofe/alist/model"

func isSharePoint(account *model.Account) bool {
	return account.InternalType == "sharepoint"
}
