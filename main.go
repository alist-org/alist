package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/public"
	"github.com/Xhofe/alist/server"
	"github.com/Xhofe/alist/utils"
	"github.com/allegro/bigcache/v3"
	"github.com/eko/gocache/v2/cache"
	"github.com/eko/gocache/v2/store"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"time"
)

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

// InitConf init config
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

func initCache() {
	log.Infof("init cache...")
	bigCacheConfig := bigcache.DefaultConfig(60 * time.Minute)
	bigCacheConfig.HardMaxCacheSize = 512
	bigCacheClient, _ := bigcache.NewBigCache(bigCacheConfig)
	bigCacheStore := store.NewBigcache(bigCacheClient, nil)
	conf.Cache = cache.New(bigCacheStore)
}


func init() {
	flag.StringVar(&conf.ConfigFile, "conf", "config.json", "config file")
	flag.BoolVar(&conf.Debug,"debug",false,"start with debug mode")
	flag.Parse()
	initLog()
	initConf()
	model.InitModel()
	initCache()
}

func main() {
	app := fiber.New()
	app.Use("/",filesystem.New(filesystem.Config{
		Root:         http.FS(public.Public),
		//NotFoundFile: "index.html",
	}))
	server.InitApiRouter(app)
	log.Info("starting server")
	_ = app.Listen(fmt.Sprintf(":%d", conf.Conf.Port))
}
