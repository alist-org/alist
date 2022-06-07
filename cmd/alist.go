package main

import (
	"flag"
	"fmt"
	"github.com/alist-org/alist/v3/bootstrap"
	"github.com/alist-org/alist/v3/cmd/args"
	"github.com/alist-org/alist/v3/conf"
	_ "github.com/alist-org/alist/v3/drivers"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"os"
)

func init() {
	flag.StringVar(&args.Config, "conf", "data/config.json", "config file")
	flag.BoolVar(&args.Debug, "debug", false, "start with debug mode")
	flag.BoolVar(&args.Version, "version", false, "print version info")
	flag.BoolVar(&args.Password, "password", false, "print current password")
	flag.BoolVar(&args.NoPrefix, "no-prefix", false, "disable env prefix")
	flag.Parse()
}

func Init() {
	if args.Version {
		fmt.Printf("Built At: %s\nGo Version: %s\nAuthor: %s\nCommit ID: %s\nVersion: %s\nWebVersion: %s\n",
			conf.BuiltAt, conf.GoVersion, conf.GitAuthor, conf.GitCommit, conf.Version, conf.WebVersion)
		os.Exit(0)
	}
	bootstrap.InitConfig()
	bootstrap.Log()
}
func main() {
	Init()
	if !args.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.LoggerWithWriter(log.StandardLogger().Out), gin.RecoveryWithWriter(log.StandardLogger().Out))
	// TODO: setup router
	base := fmt.Sprintf("%s:%d", conf.Conf.Address, conf.Conf.Port)
	log.Infof("start server @ %s", base)
	var err error
	if conf.Conf.Scheme.Https {
		err = r.RunTLS(base, conf.Conf.Scheme.CertFile, conf.Conf.Scheme.KeyFile)
	} else {
		err = r.Run(base)
	}
	if err != nil {
		log.Errorf("failed to start: %s", err.Error())
	}
}
