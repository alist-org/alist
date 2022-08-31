package local

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/operations"
)

type Addition struct {
	driver.RootFolderId
	RefreshToken   string `json:"refresh_token" required:"true"`
	OrderBy        string `json:"order_by" type:"select" options:"name,size,updated_at,created_at"`
	OrderDirection string `json:"order_direction" type:"select" options:"ASC,DESC"`
	RapidUpload    bool   `json:"rapid_upload"`
}

var config = driver.Config{
	Name:        "aliyundrive",
	DefaultRoot: "root",
}

func New() driver.Driver {
	return &AliDrive{}
}

func init() {
	operations.RegisterDriver(config, New)
}
