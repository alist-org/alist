package bootstrap

import (
	"context"
	"time"

	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/search"
	"github.com/alist-org/alist/v3/internal/setting"
	"github.com/alist-org/alist/v3/pkg/cron"
	log "github.com/sirupsen/logrus"
)

func InitIndex() {
	progress, err := search.Progress()
	if err != nil {
		log.Errorf("init index error: %+v", err)
		return
	}
	if !progress.IsDone {
		progress.IsDone = true
		search.WriteProgress(progress)
	}
	if interval := setting.GetInt(conf.SearchIndex, 0); interval > 0 {
		updateCron := cron.NewCron(time.Minute * time.Duration(interval))
		updateCron.Do(func() {
			if search.Running.Load() {
				log.Warn("auto update index failed, index is already running")
				return
			}
			indexPaths := search.GetIndexPaths()
			for _, path := range indexPaths {
				err = search.Del(context.Background(), path)
				if err != nil {
					log.Errorf("delete index on %s error: %+v", path, err)
					return
				}
			}
			log.Info("auto update index start")
			err = search.BuildIndex(context.Background(), indexPaths, search.GetIndexPaths(), -1, true)
		})
	}
}
