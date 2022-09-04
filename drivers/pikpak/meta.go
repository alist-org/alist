package pikpak

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	driver.RootID
	Username string `json:"username" required:"true"`
	Password string `json:"password" required:"true"`
}

var config = driver.Config{
	Name:        "PikPak",
	LocalSort:   true,
	DefaultRoot: "",
}

func New() driver.Driver {
	return &PikPak{}
}

func init() {
	op.RegisterDriver(config, New)
}
