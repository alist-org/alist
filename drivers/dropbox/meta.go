package dropbox

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

const (
	DefaultClientID = "76lrwrklhdn1icb"
)

type Addition struct {
	RefreshToken string `json:"refresh_token" required:"true"`
	driver.RootPath

	OauthTokenURL string `json:"oauth_token_url" default:"https://api.xhofe.top/alist/dropbox/token"`
	ClientID      string `json:"client_id" required:"false" help:"Keep it empty if you don't have one"`
	ClientSecret  string `json:"client_secret" required:"false" help:"Keep it empty if you don't have one"`

	AccessToken     string
	RootNamespaceId string
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
	NoOverwriteUpload: true,
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &Dropbox{
			base:        "https://api.dropboxapi.com",
			contentBase: "https://content.dropboxapi.com",
		}
	})
}
