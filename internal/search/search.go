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

var instance searcher.Searcher = nil

// Init or reset index
func Init(mode string) error {
	if instance != nil {
		err := instance.Release(context.Background())
		if err != nil {
			log.Errorf("release instance err: %+v", err)
		}
		instance = nil
	}
	if Running {
		return fmt.Errorf("index is running")
	}
	if mode == "none" {
		log.Warnf("not enable search")
		return nil
	}
	s, ok := searcher.NewMap[mode]
	if !ok {
		return fmt.Errorf("not support index: %s", mode)
	}
	i, err := s()
	if err != nil {
		log.Errorf("init searcher error: %+v", err)
	} else {
		instance = i
	}
	return err
}

func Search(ctx context.Context, req model.SearchReq) ([]model.SearchNode, int64, error) {
	return instance.Search(ctx, req)
}

func init() {
	db.RegisterSettingItemHook(conf.SearchIndex, func(item *model.SettingItem) error {
		log.Debugf("searcher init, mode: %s", item.Value)
		return Init(item.Value)
	})
}
