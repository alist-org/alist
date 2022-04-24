package webdav

import (
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/drivers/webdav/odrvcookie"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/pkg/gowebdav"
	"github.com/Xhofe/alist/utils"
	"net/http"
	"strings"
)

func (driver WebDav) NewClient(account *model.Account) *gowebdav.Client {
	c := gowebdav.NewClient(account.SiteUrl, account.Username, account.Password)
	if isSharePoint(account) {
		cookie, err := odrvcookie.GetCookie(account.Username, account.Password, account.SiteUrl)
		if err == nil {
			c.SetInterceptor(func(method string, rq *http.Request) {
				rq.Header.Del("Authorization")
				rq.Header.Set("Cookie", cookie)
			})
		}
	}
	return c
}

func (driver WebDav) WebDavPath(path string) string {
	path = utils.ParsePath(path)
	path = strings.TrimPrefix(path, "/")
	return path
}

func init() {
	base.RegisterDriver(&WebDav{})
}
