package local

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	driver.RootPath
	Thumbnail        bool   `json:"thumbnail" required:"true" help:"enable thumbnail"`
	ThumbCacheFolder string `json:"thumb_cache_folder"`
	ShowHidden       bool   `json:"show_hidden" default:"true" required:"false" help:"show hidden directories and files"`
	MkdirPerm        string `json:"mkdir_perm" default:"777"`
}

var config = driver.Config{
	Name:        "Local",
	OnlyLocal:   true,
	LocalSort:   true,
	NoCache:     true,
	DefaultRoot: "/",
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &Local{}
	})
}
