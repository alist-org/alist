package local

import "github.com/alist-org/alist/v3/internal/driver"

type Addition struct {
	RootFolder string `json:"root_folder" type:"string" help:"root folder path" default:"/"`
}

var config = driver.Config{
	Name:      "Local",
	OnlyLocal: true,
	LocalSort: true,
}

func New() driver.Driver {
	return &Driver{}
}

func init() {
	driver.RegisterDriver(config, New)
}
