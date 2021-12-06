package bootstrap

import (
	"github.com/Xhofe/alist/conf"
	log "github.com/sirupsen/logrus"
)

// InitLog init log
func InitLog() {
	if conf.Debug {
		log.SetLevel(log.DebugLevel)
		log.SetReportCaller(true)
	}
	log.SetFormatter(&log.TextFormatter{
		//DisableColors: true,
		ForceColors:               true,
		EnvironmentOverrideColors: true,
		TimestampFormat:           "2006-01-02 15:04:05",
		FullTimestamp:             true,
	})
	log.Infof("init log...")
}

func init() {
	InitLog()
}