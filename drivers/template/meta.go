package template

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	// Usually one of two
	driver.RootFolderPath
	driver.RootFolderId
	// define other
	Field string `json:"field" type:"select" required:"true" options:"a,b,c" default:"a"`
}

var config = driver.Config{
	Name:        "Template",
	LocalSort:   false,
	OnlyLocal:   false,
	OnlyProxy:   false,
	NoCache:     false,
	NoUpload:    false,
	NeedMs:      false,
	DefaultRoot: "root, / or other",
}

func New() driver.Driver {
	return &Template{}
}

func init() {
	op.RegisterDriver(config, New)
}
