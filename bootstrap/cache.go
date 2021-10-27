package bootstrap

import (
	"github.com/Xhofe/alist/conf"
	"github.com/allegro/bigcache/v3"
	"github.com/eko/gocache/v2/cache"
	"github.com/eko/gocache/v2/store"
	log "github.com/sirupsen/logrus"
	"time"
)

// InitCache init cache
func InitCache() {
	log.Infof("init cache...")
	bigCacheConfig := bigcache.DefaultConfig(60 * time.Minute)
	bigCacheConfig.HardMaxCacheSize = 512
	bigCacheClient, _ := bigcache.NewBigCache(bigCacheConfig)
	bigCacheStore := store.NewBigcache(bigCacheClient, nil)
	conf.Cache = cache.New(bigCacheStore)
}