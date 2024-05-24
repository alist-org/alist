package lark

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	// Usually one of two
	driver.RootPath
	// define other
	AppId           string `json:"app_id" type:"text" help:"app id"`
	AppSecret       string `json:"app_secret" type:"text" help:"app secret"`
	ExternalMode    bool   `json:"external_mode" type:"bool" help:"external mode"`
	TenantUrlPrefix string `json:"tenant_url_prefix" type:"text" help:"tenant url prefix"`
}

var config = driver.Config{
	Name:              "Lark",
	LocalSort:         false,
	OnlyLocal:         false,
	OnlyProxy:         false,
	NoCache:           false,
	NoUpload:          false,
	NeedMs:            false,
	DefaultRoot:       "/",
	CheckStatus:       false,
	Alert:             "",
	NoOverwriteUpload: true,
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &Lark{}
	})
}
