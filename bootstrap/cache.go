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
	c := conf.Conf.Cache
	if c.Expiration == 0 {
		c.Expiration, c.CleanupInterval = 60, 120
	}
	goCacheClient := goCache.New(time.Duration(c.Expiration)*time.Minute, time.Duration(c.CleanupInterval)*time.Minute)
	goCacheStore := store.NewGoCache(goCacheClient, nil)
	conf.Cache = cache.New(goCacheStore)
}
