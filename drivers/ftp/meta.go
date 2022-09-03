package ftp

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	Address  string `json:"address" required:"true"`
	Username string `json:"username" required:"true"`
	Password string `json:"password" required:"true"`
	driver.RootFolderPath
}

var config = driver.Config{
	Name:        "FTP",
	LocalSort:   false,
	OnlyLocal:   false,
	OnlyProxy:   false,
	NoCache:     false,
	NoUpload:    false,
	NeedMs:      false,
	DefaultRoot: "root, / or other",
}

func New() driver.Driver {
	return &FTP{}
}

func init() {
	op.RegisterDriver(config, New)
}
