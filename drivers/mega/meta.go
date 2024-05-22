package mega

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	// Usually one of two
	//driver.RootPath
	//driver.RootID
	Email       string `json:"email" required:"true"`
	Password    string `json:"password" required:"true"`
	TwoFACode   string `json:"two_fa_code" required:"false" help:"2FA 6-digit code, filling in the 2FA code alone will not support reloading driver"`
	TwoFASecret string `json:"two_fa_secret" required:"false" help:"2FA secret"`
}

var config = driver.Config{
	Name:      "Mega_nz",
	LocalSort: true,
	OnlyLocal: true,
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &Mega{}
	})
}
