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

func Init() {
	flag.StringVar(&conf.ConfigFile, "conf", "config.json", "config file")
	flag.BoolVar(&conf.Debug,"debug",false,"start with debug mode")
	flag.Parse()
	bootstrap.InitLog()
	bootstrap.InitConf()
	bootstrap.InitCron()
	bootstrap.InitModel()
	bootstrap.InitCache()
}

func main() {
	Init()
	app := fiber.New()
	server.InitApiRouter(app)
	app.Use("/",filesystem.New(filesystem.Config{
		Root:         http.FS(public.Public),
		NotFoundFile: "index.html",
	}))
	log.Info("starting server")
	err := app.Listen(fmt.Sprintf(":%d", conf.Conf.Port))
	if err != nil {
		log.Errorf("failed to start: %s", err.Error())
	}
}
