package pikpak

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	driver.RootID
	Username           string `json:"username" required:"true"`
	Password           string `json:"password" required:"true"`
	Platform           string `json:"platform" required:"true" type:"select" options:"android,web"`
	RefreshToken       string `json:"refresh_token" required:"true" default:""`
	RefreshTokenMethod string `json:"refresh_token_method" required:"true" type:"select" options:"oauth2,http"`
	CaptchaToken       string `json:"captcha_token" default:""`
	DeviceID           string `json:"device_id"  required:"false" default:""`
	DisableMediaLink   bool   `json:"disable_media_link" default:"true"`
}

var config = driver.Config{
	Name:        "PikPak",
	LocalSort:   true,
	DefaultRoot: "",
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &PikPak{}
	})
}
