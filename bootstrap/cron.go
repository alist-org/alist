package bootstrap

import (
	"github.com/Xhofe/alist/conf"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
)

// InitCron init cron
func InitCron() {
	log.Infof("init cron...")
	conf.Cron = cron.New()
	conf.Cron.Start()
}
