package seafile

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	driver.RootPath

	Address  string `json:"address" required:"true"`
	UserName string `json:"username" required:"true"`
	Password string `json:"password" required:"true"`
	RepoId   string `json:"repoId" required:"true"`
}

var config = driver.Config{
	Name:        "Seafile",
	DefaultRoot: "/",
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &Seafile{}
	})
}
