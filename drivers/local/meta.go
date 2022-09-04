package local

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	driver.RootPath
	Thumbnail bool `json:"thumbnail" required:"true" help:"enable thumbnail"`
}

var config = driver.Config{
	Name:        "Local",
	OnlyLocal:   true,
	LocalSort:   true,
	NoCache:     true,
	DefaultRoot: "/",
}

func New() driver.Driver {
	return &Local{}
}

func init() {
	op.RegisterDriver(config, New)
}
