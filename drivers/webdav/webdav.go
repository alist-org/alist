package webdav

import (
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/drivers/webdav/odrvcookie"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	log "github.com/sirupsen/logrus"
	"github.com/studio-b12/gowebdav"
	"net/http"
	"strings"
)

func (driver WebDav) NewClient(account *model.Account) *gowebdav.Client {
	c := gowebdav.NewClient(account.SiteUrl, account.Username, account.Password)
	if isSharePoint(account) {
		cookie, err := odrvcookie.GetCookie(account.Username, account.Password, account.SiteUrl)
		log.Debugln(cookie, err)
		if err == nil {
			log.Debugln("set interceptor")
			c.SetInterceptor(func(method string, rq *http.Request) {
				rq.Header.Del("Authorization")
				rq.Header.Set("Cookie", cookie)
				log.Debugf("sp webdav req: %+v", rq)
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
