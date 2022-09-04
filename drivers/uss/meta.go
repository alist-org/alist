package uss

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	driver.RootPath
	Bucket           string `json:"bucket" required:"true"`
	Endpoint         string `json:"endpoint" required:"true"`
	OperatorName     string `json:"operator_name" required:"true"`
	OperatorPassword string `json:"operator_password" required:"true"`
	CustomHost       string `json:"custom_host"`
	SignURLExpire    int    `json:"sign_url_expire" type:"number" default:"4"`
}

var config = driver.Config{
	Name:      "USS",
	LocalSort: true,
}

func New() driver.Driver {
	return &USS{}
}

func init() {
	op.RegisterDriver(config, New)
}
