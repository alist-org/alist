package aliyundrive_open

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	driver.RootID
	RefreshToken   string `json:"refresh_token" required:"true"`
	OrderBy        string `json:"order_by" type:"select" options:"name,size,updated_at,created_at"`
	OrderDirection string `json:"order_direction" type:"select" options:"ASC,DESC"`
	OauthTokenURL  string `json:"oauth_token_url" default:"https://api.nn.ci/alist/ali_open/token"`
	ClientID       string `json:"client_id" required:"false"`
	ClientSecret   string `json:"client_secret" required:"false"`
}

var config = driver.Config{
	Name:        "AliyundriveOpen",
	LocalSort:   false,
	OnlyLocal:   false,
	OnlyProxy:   false,
	NoCache:     false,
	NoUpload:    false,
	NeedMs:      false,
	DefaultRoot: "root",
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &AliyundriveOpen{
			base: "https://open.aliyundrive.com",
		}
	})
}
