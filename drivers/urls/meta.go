package urls

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	// Usually one of two
	// driver.RootPath
	// driver.RootID
	// define other
	UrlStructure string `json:"url_structure" type:"text" required:"true" default:"https://jsd.nn.ci/gh/alist-org/alist/README.md\nhttps://jsd.nn.ci/gh/alist-org/alist/README_cn.md\nfolder:\n  https://jsd.nn.ci/gh/alist-org/alist/CONTRIBUTING.md\n  https://jsd.nn.ci/gh/alist-org/alist/CODE_OF_CONDUCT.md.md"`
}

var config = driver.Config{
	Name:              "Urls",
	LocalSort:         true,
	OnlyLocal:         false,
	OnlyProxy:         false,
	NoCache:           true,
	NoUpload:          true,
	NeedMs:            false,
	DefaultRoot:       "",
	CheckStatus:       false,
	Alert:             "",
	NoOverwriteUpload: false,
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &Urls{}
	})
}
