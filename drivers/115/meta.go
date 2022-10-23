package _115

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	// Username string `json:"username" required:"true"`
	// Password string `json:"password" required:"true"`
	UID  string `json:"uid" required:"true"`
	CID  string `json:"cid" required:"true"`
	SEID string `json:"seid" required:"true"`
	driver.RootID
}

var config = driver.Config{
	Name:        "115Cloud",
	DefaultRoot: "0",
}

func New() driver.Driver {
	return &Pan115{}
}

func init() {
	op.RegisterDriver(config, New)
}
