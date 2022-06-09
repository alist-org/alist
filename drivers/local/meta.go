package local

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/operations"
)

type Addition struct {
	RootFolder string `json:"root_folder" help:"root folder path" default:"/"`
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
	operations.RegisterDriver(config, New)
}
