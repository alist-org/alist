package webdav

import (
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/studio-b12/gowebdav"
	"strings"
)

func (driver WebDav) NewClient(account *model.Account) *gowebdav.Client {
	return gowebdav.NewClient(account.SiteUrl, account.Username, account.Password)
}

func (driver WebDav) WebDavPath(path string) string {
	path = utils.ParsePath(path)
	path = strings.TrimPrefix(path, "/")
	return path
}

func init() {
	base.RegisterDriver(&WebDav{})
}
