package _139

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	//Account       string `json:"account" required:"true"`
	Authorization string `json:"authorization" type:"text" required:"true"`
	driver.RootID
	Type    string `json:"type" type:"select" options:"personal,family" default:"personal"`
	CloudID string `json:"cloud_id"`
}

var config = driver.Config{
	Name:      "139Yun",
	LocalSort: true,
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &Yun139{}
	})
}
