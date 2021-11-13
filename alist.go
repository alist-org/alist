package main

import (
	"flag"
	"fmt"
	"github.com/Xhofe/alist/bootstrap"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/public"
	"github.com/Xhofe/alist/server"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	log "github.com/sirupsen/logrus"
	"net/http"
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
	app := fiber.New()
	server.InitApiRouter(app)
	app.Use("/", filesystem.New(filesystem.Config{
		Root:         http.FS(public.Public),
		NotFoundFile: "index.html",
	}))
	log.Info("starting server")
	err := app.Listen(fmt.Sprintf(":%d", conf.Conf.Port))
	if err != nil {
		log.Errorf("failed to start: %s", err.Error())
	}
}
