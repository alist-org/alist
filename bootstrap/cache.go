package bootstrap

import (
	"github.com/Xhofe/alist/conf"
	"github.com/eko/gocache/v2/cache"
	"github.com/eko/gocache/v2/store"
	goCache "github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
	"time"
)

// InitCache init cache
func InitCache() {
	log.Infof("init cache...")
	goCacheClient := goCache.New(60*time.Minute, 120*time.Minute)
	goCacheStore := store.NewGoCache(goCacheClient, nil)
	conf.Cache = cache.New(goCacheStore)
}
