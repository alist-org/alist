package fastwebdav

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	driver.RootPath
	Address string `json:"address" required:"true"`
	APIKey  string `json:"apikey" required:"true"`
}

var config = driver.Config{
	Name:        "FastWebdav",
	DefaultRoot: "/",
	NoUpload:    true,
	Alert:       "只支持文件读取",
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &FastWebdav{}
	})
}
