package _189pc

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	Username string `json:"username" required:"true"`
	Password string `json:"password" required:"true"`
	VCode    string `json:"validate_code"`
	driver.RootID
	OrderBy        string `json:"order_by" type:"select" options:"filename,filesize,lastOpTime" default:"filename"`
	OrderDirection string `json:"order_direction" type:"select" options:"asc,desc" default:"asc"`
	Type           string `json:"type" type:"select" options:"personal,family" default:"personal"`
	FamilyID       string `json:"family_id"`
	RapidUpload    bool   `json:"rapid_upload"`
}

var config = driver.Config{
	Name:        "189CloudPC",
	DefaultRoot: "-11",
}

func init() {
	op.RegisterDriver(config, func() driver.Driver {
		return &Yun189PC{}
	})
}
