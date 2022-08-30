package onedrive

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/operations"
)

type Addition struct {
	driver.RootFolderPath
	Region       string `json:"region" type:"select" required:"true" options:"global,cn,us,de"`
	IsSharepoint bool   `json:"is_sharepoint"`
	ClientId     string `json:"client_id" required:"true"`
	ClientSecret string `json:"client_secret" required:"true"`
	RedirectUri  string `json:"redirect_uri" required:"true" default:"https://tool.nn.ci/onedrive/callback"`
	RefreshToken string `json:"refresh_token" required:"true"`
	SiteId       string `json:"site_id"`
}

var config = driver.Config{
	Name:        "onedrive",
	LocalSort:   true,
	DefaultRoot: "/",
}

func New() driver.Driver {
	return &Onedrive{}
}

func init() {
	operations.RegisterDriver(config, New)
}
