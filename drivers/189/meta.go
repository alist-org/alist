package _189

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	Username string `json:"username" required:"true"`
	Password string `json:"password" required:"true"`
	driver.RootID
}

var config = driver.Config{
	Name:        "189Cloud",
	LocalSort:   true,
	DefaultRoot: "-11",
}

func init() {
	op.RegisterDriver(config, func() driver.Driver {
		return &Cloud189{}
	})
}
