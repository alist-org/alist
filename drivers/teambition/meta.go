package teambition

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	Region    string `json:"region" type:"select" options:"china,international" required:"true"`
	Cookie    string `json:"cookie" required:"true"`
	ProjectID string `json:"project_id" required:"true"`
	driver.RootID
	OrderBy           string `json:"order_by" type:"select" options:"fileName,fileSize,updated,created" default:"fileName"`
	OrderDirection    string `json:"order_direction" type:"select" options:"Asc,Desc" default:"Asc"`
	UseS3UploadMethod bool   `json:"use_s3_upload_method" default:"true"`
}

var config = driver.Config{
	Name: "Teambition",
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &Teambition{}
	})
}
