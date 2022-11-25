package search

import (
	"github.com/alist-org/alist/v3/internal/search/searcher"
)

type New func() searcher.Searcher

var searcherNewMap = map[string]New{}

func RegisterDriver(config searcher.Config, searcher New) {
	searcherNewMap[config.Name] = searcher
}
