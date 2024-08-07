package quark_uc_tv

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	// Usually one of two
	driver.RootID
	// define other
	RefreshToken string `json:"refresh_token" required:"false" default:""`
	// 必要且影响登录,由签名决定
	DeviceID string `json:"device_id"  required:"false" default:""`
	// 登陆所用的数据 无需手动填写
	QueryToken string `json:"query_token" required:"false" default:"" help:"don't edit'"`
}

type Conf struct {
	api      string
	clientID string
	signKey  string
	appVer   string
	channel  string
	codeApi  string
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &QuarkUCTV{
			config: driver.Config{
				Name:              "QuarkTV",
				OnlyLocal:         false,
				DefaultRoot:       "0",
				NoOverwriteUpload: true,
				NoUpload:          true,
			},
			conf: Conf{
				api:      "https://open-api-drive.quark.cn",
				clientID: "d3194e61504e493eb6222857bccfed94",
				signKey:  "kw2dvtd7p4t3pjl2d9ed9yc8yej8kw2d",
				appVer:   "1.5.6",
				channel:  "CP",
				codeApi:  "http://api.extscreen.com/quarkdrive",
			},
		}
	})
	op.RegisterDriver(func() driver.Driver {
		return &QuarkUCTV{
			config: driver.Config{
				Name:              "UCTV",
				OnlyLocal:         false,
				DefaultRoot:       "0",
				NoOverwriteUpload: true,
				NoUpload:          true,
			},
			conf: Conf{
				api:      "https://open-api-drive.uc.cn",
				clientID: "5acf882d27b74502b7040b0c65519aa7",
				signKey:  "l3srvtd7p42l0d0x1u8d7yc8ye9kki4d",
				appVer:   "1.6.5",
				channel:  "UCTVOFFICIALWEB",
				codeApi:  "http://api.extscreen.com/ucdrive",
			},
		}
	})
}
