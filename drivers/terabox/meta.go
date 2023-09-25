package terabox

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	driver.RootPath
	Cookie string `json:"cookie" required:"true"`
	//JsToken        string `json:"js_token" type:"string" required:"true"`
	DownloadAPI    string `json:"download_api" type:"select" options:"official,crack" default:"official"`
	OrderBy        string `json:"order_by" type:"select" options:"name,time,size" default:"name"`
	OrderDirection string `json:"order_direction" type:"select" options:"asc,desc" default:"asc"`
}

var config = driver.Config{
	Name:        "Terabox",
	DefaultRoot: "/",
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &Terabox{}
	})
}
