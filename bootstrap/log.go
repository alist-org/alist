package bootstrap

import (
	"flag"
	"github.com/Xhofe/alist/conf"
	log "github.com/sirupsen/logrus"
)

// InitLog init log
func InitLog() {
	if conf.Debug {
		log.SetLevel(log.DebugLevel)
		log.SetReportCaller(true)
	}
	if conf.Password || conf.Version {
		log.SetLevel(log.WarnLevel)
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
	flag.StringVar(&conf.ConfigFile, "conf", "data/config.json", "config file")
	flag.BoolVar(&conf.Debug, "debug", false, "start with debug mode")
	flag.BoolVar(&conf.Version, "version", false, "print version info")
	flag.BoolVar(&conf.Password, "password", false, "print current password")
	flag.BoolVar(&conf.Docker, "docker", false, "is using docker")
	flag.Parse()
	InitLog()
}
