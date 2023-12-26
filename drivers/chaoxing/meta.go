package chaoxing

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

// 此程序挂载的是超星小组网盘，需要代理才能使用；
// 登录超星后进入个人空间，进入小组，新建小组，点击进去。
// url中就有bbsid的参数，系统限制单文件大小2G，没有总容量限制
type Addition struct {
	// 超星用户名及密码
	UserName string `json:"user_name" required:"true"`
	Password string `json:"password" required:"true"`
	// 从自己新建的小组url里获取
	Bbsid string `json:"bbsid" required:"true"`
	driver.RootID
	// 可不填，程序会自动登录获取
	Cookie string `json:"cookie"`
}

type Conf struct {
	ua         string
	referer    string
	api        string
	DowloadApi string
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &ChaoXing{
			config: driver.Config{
				Name:              "ChaoXingGroupDrive",
				OnlyProxy:         true,
				OnlyLocal:         false,
				DefaultRoot:       "-1",
				NoOverwriteUpload: true,
			},
			conf: Conf{
				ua:         "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) quark-cloud-drive/2.5.20 Chrome/100.0.4896.160 Electron/18.3.5.4-b478491100 Safari/537.36 Channel/pckk_other_ch",
				referer:    "https://chaoxing.com/",
				api:        "https://groupweb.chaoxing.com",
				DowloadApi: "https://noteyd.chaoxing.com",
			},
		}
	})
}
