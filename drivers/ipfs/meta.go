package ipfs

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	// Usually one of two
	driver.RootPath
	Endpoint string `json:"endpoint"`
	Gateway  string `json:"gateway" default:"https://ipfs.io"`
}

var config = driver.Config{
	Name:        "IPFS",
	DefaultRoot: "root",
	// NoUpload:    true,
	LocalSort: true,
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &IPFS{}
	})
}
