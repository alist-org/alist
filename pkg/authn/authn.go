package authn

import (
	"net/http"
	"net/url"

	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/setting"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/go-webauthn/webauthn/webauthn"
)

func NewAuthnInstance(r *http.Request) (*webauthn.WebAuthn, error) {
	siteurl, err := url.Parse(common.GetApiUrl(r))
	if err != nil {
		return nil, err
	}
	return webauthn.New(&webauthn.Config{
		RPDisplayName: setting.GetStr(conf.SiteTitle),
		RPID:          siteurl.Hostname(),
		RPOrigin:      siteurl.String(),
		// RPOrigin: "http://localhost:5173"
	})
}

type Authn struct{}

func (a *Authn) UpdateAuthn(userID uint, authn string) error {
	return db.UpdateAuthn(userID, authn)
}
