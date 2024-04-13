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
	Alert:       "warning|只能读文件不能进行其它操作,如:复制,移动,上传等",
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &FastWebdav{}
	})
}
