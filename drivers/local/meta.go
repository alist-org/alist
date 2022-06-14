package local

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/operations"
)

type Addition struct {
	driver.RootFolderPath
}

var config = driver.Config{
	Name:      "Local",
	OnlyLocal: true,
	LocalSort: true,
	NoCache:   true,
}

func New() driver.Driver {
	return &Driver{}
}

func init() {
	operations.RegisterDriver(config, New)
}
