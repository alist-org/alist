package vtencent

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	driver.RootID
	Cookie         string `json:"cookie" required:"true"`
	TfUid          string `json:"tf_uid"`
	OrderBy        string `json:"order_by" type:"select" options:"Name,Size,UpdateTime,CreatTime"`
	OrderDirection string `json:"order_direction" type:"select" options:"Asc,Desc"`
}

type Conf struct {
	ua      string
	referer string
	origin  string
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &Vtencent{
			config: driver.Config{
				Name:              "VTencent",
				OnlyProxy:         true,
				OnlyLocal:         false,
				DefaultRoot:       "9",
				NoOverwriteUpload: true,
			},
			conf: Conf{
				ua:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) quark-cloud-drive/2.5.20 Chrome/100.0.4896.160 Electron/18.3.5.4-b478491100 Safari/537.36 Channel/pckk_other_ch",
				referer: "https://app.v.tencent.com/",
				origin:  "https://app.v.tencent.com",
			},
		}
	})
}
