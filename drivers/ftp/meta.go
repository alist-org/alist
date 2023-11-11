package ftp

import (
	"github.com/alist-org/alist/v3/internal2/driver"
	"github.com/alist-org/alist/v3/internal2/op"
)

type Addition struct {
	Address  string `json:"address" required:"true"`
	Username string `json:"username" required:"true"`
	Password string `json:"password" required:"true"`
	driver.RootPath
}

var config = driver.Config{
	Name:        "FTP",
	LocalSort:   true,
	OnlyLocal:   true,
	DefaultRoot: "/",
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &FTP{}
	})
}
