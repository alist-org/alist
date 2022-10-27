package alist_v3

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	driver.RootPath
	Address     string `json:"url" required:"true"`
	Password    string `json:"password"`
	AccessToken string `json:"access_token"`
}

var config = driver.Config{
	Name:        "AList V3",
	LocalSort:   true,
	NoUpload:    true,
	DefaultRoot: "/",
}

func init() {
	op.RegisterDriver(config, func() driver.Driver {
		return &AListV3{}
	})
}
