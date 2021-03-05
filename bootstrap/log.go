package bootstrap

import (
	"github.com/Xhofe/alist/conf"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// init logrus
func InitLog() {
	if conf.Debug {
		log.SetLevel(log.DebugLevel)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	log.SetFormatter(&log.TextFormatter{
		ForceColors:               true,
		EnvironmentOverrideColors: true,
		TimestampFormat:           "2006-01-02 15:04:05",
		FullTimestamp:             true,
	})
}
