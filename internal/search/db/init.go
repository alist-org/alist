package db

import (
	"github.com/alist-org/alist/v3/internal/search/searcher"
)

var config = searcher.Config{
	Name:       "database",
	AutoUpdate: true,
}

func init() {
	searcher.RegisterSearcher(config, func() (searcher.Searcher, error) {
		return &DB{}, nil
	})
}
