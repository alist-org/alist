package bootstrap

import (
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/search/bleve"
)

func InitIndex() {
	bleve.Init(&conf.Conf.IndexDir)
}
