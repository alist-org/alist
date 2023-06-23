package template

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	// Usually one of two
	driver.RootPath
	driver.RootID
	// define other
	RefreshToken  string `json:"refresh_token" required:"true"`
	OauthTokenURL string `json:"oauth_token_url" default:"https://api.xhofe.top/alist/dropbox/token"`
	ClientID      string `json:"client_id" required:"false" help:"Keep it empty if you don't have one"`
	ClientSecret  string `json:"client_secret" required:"false" help:"Keep it empty if you don't have one"`

	AccessToken string
}

var config = driver.Config{
	Name:              "Dropbox",
	LocalSort:         false,
	OnlyLocal:         false,
	OnlyProxy:         false,
	NoCache:           false,
	NoUpload:          false,
	NeedMs:            false,
	DefaultRoot:       "",
	CheckStatus:       false,
	Alert:             "",
	NoOverwriteUpload: false,
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &Dropbox{}
	})
}
