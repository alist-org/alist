package main

import (
	"flag"
	"fmt"
	"github.com/Xhofe/alist/bootstrap"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/server"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func init() {
	flag.StringVar(&conf.ConfigFile, "conf", "data/config.json", "config file")
	flag.BoolVar(&conf.Debug, "debug", false, "start with debug mode")
	flag.BoolVar(&conf.Version, "version", false, "print version info")
	flag.Parse()
}

func Init() {
	bootstrap.InitLog()
	bootstrap.InitConf()
	bootstrap.InitCron()
	bootstrap.InitModel()
	bootstrap.InitCache()
}

func main() {
	if conf.Version {
		fmt.Printf("Built At: %s\nGo Version: %s\nAuthor: %s\nCommit ID: %s\nVersion: %s\n", conf.BuiltAt, conf.GoVersion, conf.GitAuthor, conf.GitCommit, conf.GitTag)
		return
	}
	Init()
	if !conf.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	server.InitApiRouter(r)

	log.Info("starting server")
	err := r.Run(fmt.Sprintf("%s:%d", conf.Conf.Address, conf.Conf.Port))
	if err != nil {
		log.Errorf("failed to start: %s", err.Error())
	}
}
