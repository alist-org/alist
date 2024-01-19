package quqi

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	driver.RootID
	Phone    string `json:"phone"`
	Password string `json:"password"`
	Cookie   string `json:"cookie" help:"Cookie can be used on multiple clients at the same time"`
}

var config = driver.Config{
	Name:      "Quqi",
	OnlyLocal: true,
	LocalSort: true,
	//NoUpload:    true,
	DefaultRoot: "0",
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &Quqi{}
	})
}
