package yandex_disk

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	RefreshToken   string `json:"refresh_token" required:"true"`
	OrderBy        string `json:"order_by" type:"select" options:"name,path,created,modified,size" default:"name"`
	OrderDirection string `json:"order_direction" type:"select" options:"asc,desc" default:"asc"`
	driver.RootPath
	ClientID     string `json:"client_id" required:"true" default:"a78d5a69054042fa936f6c77f9a0ae8b"`
	ClientSecret string `json:"client_secret" required:"true" default:"9c119bbb04b346d2a52aa64401936b2b"`
}

var config = driver.Config{
	Name:        "YandexDisk",
	DefaultRoot: "/",
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &YandexDisk{}
	})
}
