package webdav

import (
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/drivers/webdav/odrvcookie"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/Xhofe/alist/utils/cookie"
	log "github.com/sirupsen/logrus"
	"github.com/studio-b12/gowebdav"
	"net/http"
	"strings"
)

func (driver WebDav) NewClient(account *model.Account) *gowebdav.Client {
	c := gowebdav.NewClient(account.SiteUrl, account.Username, account.Password)
	if account.InternalType == "sharepoint" {

		ca := odrvcookie.New(account.Username, account.Password, account.SiteUrl)
		tokenConf, err := ca.Cookies()
		log.Debugln(err, tokenConf)
		if err != nil {
			c.SetInterceptor(func(method string, rq *http.Request) {
				rq.Header.Del("Authorization")
				//rq.AddCookie(&tokenConf.FedAuth)
				//rq.AddCookie(&tokenConf.RtFa)
				rq.Header.Set("Cookie", cookie.ToString([]*http.Cookie{&tokenConf.RtFa, &tokenConf.FedAuth}))
				if method == "PROPFIND" {
					rq.Header.Set("Depth", "0")
				}
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
