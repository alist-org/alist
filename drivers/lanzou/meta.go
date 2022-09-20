package lanzou

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	Type   string `json:"type" type:"select" options:"cookie,url" default:"cookie"`
	Cookie string `json:"cookie" required:"true" help:"about 15 days valid"`
	driver.RootID
	SharePassword string `json:"share_password"`
	BaseUrl       string `json:"baseUrl" required:"true" default:"https://pc.woozooo.com"`
	ShareUrl      string `json:"shareUrl" required:"true" default:"https://pan.lanzouo.com"`
}

func (a *Addition) IsCookie() bool {
	return a.Type == "cookie"
}

var config = driver.Config{
	Name:        "Lanzou",
	LocalSort:   true,
	DefaultRoot: "-1",
}

func init() {
	op.RegisterDriver(config, func() driver.Driver {
		return &LanZou{}
	})
}
