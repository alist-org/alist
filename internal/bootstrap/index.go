package bootstrap

import (
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/index"
)

func InitIndex() {
	index.Init(&conf.Conf.IndexDir)
}
