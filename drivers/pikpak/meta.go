package pikpak

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	driver.RootID
	Username         string `json:"username" required:"true"`
	Password         string `json:"password" required:"true"`
	ClientID         string `json:"client_id" required:"true" default:"YNxT9w7GMdWvEOKa"`
	ClientSecret     string `json:"client_secret" required:"true" default:"dbw2OtmVEeuUvIptb1Coyg"`
	DisableMediaLink bool   `json:"disable_media_link"`
}

var config = driver.Config{
	Name:        "PikPak",
	LocalSort:   true,
	DefaultRoot: "",
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &PikPak{}
	})
}
