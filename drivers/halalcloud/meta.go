package halalcloud

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	// Usually one of two
	driver.RootPath
	// define other
	RefreshToken string `json:"refresh_token" required:"true" help:"login type is refresh_token,this is required"`
	UploadThread string `json:"upload_thread" default:"3" help:"1 <= thread <= 32"`

	AppID      string `json:"app_id" required:"true" default:"devDebugger/1.0"`
	AppVersion string `json:"app_version" required:"true" default:"1.0.0"`
	AppSecret  string `json:"app_secret" required:"true" default:"Nkx3Y2xvZ2luLmNu"`
}

var config = driver.Config{
	Name:              "HalalCloud",
	LocalSort:         false,
	OnlyLocal:         true,
	OnlyProxy:         true,
	NoCache:           false,
	NoUpload:          false,
	NeedMs:            false,
	DefaultRoot:       "/",
	CheckStatus:       false,
	Alert:             "",
	NoOverwriteUpload: false,
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &HalalCloud{}
	})
}
