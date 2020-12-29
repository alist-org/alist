package bootstrap

import (
	"github.com/Xhofe/alist/conf"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
	"time"
)

func InitCache() {
	if conf.Conf.Cache.Enable {
		log.Infof("初始化缓存...")
		conf.Cache=cache.New(time.Duration(conf.Conf.Cache.Expiration)*time.Minute,time.Duration(conf.Conf.Cache.CleanupInterval)*time.Minute)
	}
}
