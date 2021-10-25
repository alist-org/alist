package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
)

// initConf init config
func initConf() {
	log.Infof("reading config file: %s", conf.ConfigFile)
	if !utils.Exists(conf.ConfigFile) {
		log.Infof("config file not exists, creating default config file")
		conf.Conf = conf.DefaultConfig()
		if !utils.WriteToJson(conf.ConfigFile, conf.Conf) {
			log.Fatalf("failed to create default config file")
		}
		return
	}
	config, err := ioutil.ReadFile(conf.ConfigFile)
	if err != nil {
		log.Fatalf("reading config file error:%s", err.Error())
	}
	conf.Conf = new(conf.Config)
	err = json.Unmarshal(config, conf.Conf)
	if err != nil {
		log.Fatalf("load config error: %s", err.Error())
	}
	log.Debugf("config:%+v", conf.Conf)
}

// initLog init log
func initLog() {
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
}

func init() {
	flag.StringVar(&conf.ConfigFile, "conf", "config.json", "config file")
	flag.BoolVar(&conf.Debug,"debug",false,"start with debug mode")
	flag.Parse()
	initLog()
	initConf()
	model.InitModel()
}

func main() {
	app := fiber.New()
	app.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.SendString("Hello, World 👋!")
	})
	log.Info("starting server")
	_ = app.Listen(fmt.Sprintf(":%d", conf.Conf.Port))
}
