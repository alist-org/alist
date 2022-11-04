package _115

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	Cookie        string `json:"cookie"`
	QRCodeToken string `json:"qrcode_token"`
	driver.RootID
}

var config = driver.Config{
	Name:        "115 Cloud",
	DefaultRoot: "0",
	OnlyProxy:   true,
	OnlyLocal:   true,
}

func New() driver.Driver {
	return &Pan115{}
}

func init() {
	op.RegisterDriver(config, New)
}
