package search

import (
	"context"
	"fmt"

	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/search/searcher"
	log "github.com/sirupsen/logrus"
)

var instance searcher.Searcher

func Init(mode string) error {
	if instance != nil {
		_ = instance.Drop(context.Background())
		instance = nil
	}
	if mode == "none" {
		log.Warnf("not enable search")
		return nil
	}
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

func init() {
	db.RegisterSettingItemHook(conf.SearchIndex, func(item *model.SettingItem) error {
		return Init(item.Value)
	})
}
