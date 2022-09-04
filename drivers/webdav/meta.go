package webdav

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	Vendor   string `json:"vendor" type:"select" options:"sharepoint,other" default:"other"`
	Address  string `json:"address" required:"true"`
	Username string `json:"username" required:"true"`
	Password string `json:"password" required:"true"`
	driver.RootPath
}

var config = driver.Config{
	Name:        "WebDav",
	LocalSort:   true,
	OnlyLocal:   true,
	DefaultRoot: "/",
}

func New() driver.Driver {
	return &WebDav{}
}

func init() {
	op.RegisterDriver(config, New)
}
