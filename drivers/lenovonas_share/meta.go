package LenovoNasShare

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	driver.RootPath
	ShareId  string `json:"share_id" required:"true" help:"The part after the last / in the shared link"`
	SharePwd string `json:"share_pwd" required:"true" help:"The password of the shared link"`
	Host     string `json:"host" required:"true" default:"https://siot-share.lenovo.com.cn" help:"You can change it to your local area network"`
}

var config = driver.Config{
	Name:              "LenovoNasShare",
	LocalSort:         true,
	OnlyLocal:         false,
	OnlyProxy:         false,
	NoCache:           false,
	NoUpload:          true,
	NeedMs:            false,
	DefaultRoot:       "",
	CheckStatus:       false,
	Alert:             "",
	NoOverwriteUpload: false,
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &LenovoNasShare{}
	})
}
