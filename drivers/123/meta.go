package _123

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	Username       string `json:"username" required:"true"`
	Password       string `json:"password" required:"true"`
	OrderBy        string `json:"order_by" type:"select" options:"file_name,size,update_at" default:"file_name"`
	OrderDirection string `json:"order_direction" type:"select" options:"asc,desc" default:"asc"`
	driver.RootID
	// define other
	StreamUpload bool `json:"stream_upload"`
	//Field string `json:"field" type:"select" required:"true" options:"a,b,c" default:"a"`
}

var config = driver.Config{
	Name:        "123Pan",
	DefaultRoot: "0",
}

func New() driver.Driver {
	return &Pan123{}
}

func init() {
	op.RegisterDriver(config, New)
}
