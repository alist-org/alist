package search

import (
	"context"
	"fmt"

	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/fs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/search/searcher"
)

var instance searcher.Searcher

func Init(mode string) error {
	s, ok := searcher.NewMap[mode]
	if !ok {
		return fmt.Errorf("not support index: %s", mode)
	}
	instance = s()
	return nil
}

func Search(ctx context.Context, req model.SearchReq) ([]model.SearchNode, int64, error) {
	return instance.Search(ctx, req)
}

func Progress(ctx context.Context) (*model.IndexProgress, error) {
	// TODO
	panic("")
}

func BuildIndex(ctx context.Context, indexPaths, ignorePaths []string, maxDepth int) error {
	// TODO
	fs.Get(ctx, "/")
	panic("")
}

func init() {
	db.SettingItemHooks[conf.SearchIndex] = db.SettingItemHook{
		Hook: func(item *model.SettingItem) error {
			return Init(item.Value)
		}}
}
