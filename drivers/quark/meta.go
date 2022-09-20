package quark

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	Cookie string `json:"cookie" required:"true"`
	driver.RootID
	OrderBy        string `json:"order_by" type:"select" options:"none,file_type,file_name,updated_at" default:"none"`
	OrderDirection string `json:"order_direction" type:"select" options:"asc,desc" default:"asc"`
}

var config = driver.Config{
	Name:        "Quark",
	OnlyProxy:   true,
	DefaultRoot: "0",
}

func New() driver.Driver {
	return &Quark{}
}

func init() {
	op.RegisterDriver(config, New)
}
