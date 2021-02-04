package bootstrap

import (
	"github.com/Xhofe/alist/alidrive"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
)

var Cron *cron.Cron

// refresh token func for cron
func refreshToken()  {
	alidrive.RefreshToken()
}

// init cron jobs
func InitCron() {
	log.Infof("初始化定时任务:刷新token")
	Cron=cron.New()
	_,err:=Cron.AddFunc("@every 2h",refreshToken)
	if err!=nil {
		log.Errorf("添加启动任务失败:%s",err.Error())
	}
	Cron.Start()
}