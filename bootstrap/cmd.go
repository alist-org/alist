package bootstrap

import (
	"flag"
	"github.com/Xhofe/alist/conf"
	serv "github.com/Xhofe/alist/server"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func init() {
	flag.BoolVar(&conf.Debug,"debug",false,"use debug mode")
	flag.BoolVar(&conf.Help,"help",false,"show usage help")
	flag.StringVar(&conf.Con,"conf","conf.yml","config file")
}

func Run()  {
	flag.Parse()
	if conf.Help {
		flag.Usage()
		return
	}
	start()
}

func printASC()  {
	log.Info(`
 ________  ___       ___  ________  _________   
|\   __  \|\  \     |\  \|\   ____\|\___   ___\ 
\ \  \|\  \ \  \    \ \  \ \  \___|\|___ \  \_| 
 \ \   __  \ \  \    \ \  \ \_____  \   \ \  \  
  \ \  \ \  \ \  \____\ \  \|____|\  \   \ \  \ 
   \ \__\ \__\ \_______\ \__\____\_\  \   \ \__\
    \|__|\|__|\|_______|\|__|\_________\   \|__|
                            \|_________|
`)
}

func start() {
	InitLog()
	printASC()
	if !ReadConf(conf.Con) {
		log.Errorf("读取配置文件时出现错误,启动失败.")
		return
	}
	InitClient()
	if !InitAliDrive() {
		log.Errorf("初始化阿里云盘出现错误,启动失败.")
		return
	}
	InitCron()
	server()
}

func server() {
	baseServer:="0.0.0.0:"+conf.Conf.Server.Port
	r:=gin.Default()
	serv.InitRouter(r)
	log.Infof("Starting server @ %s",baseServer)
	err:=r.Run(baseServer)
	if err!=nil {
		log.Errorf("Server failed start:%s",err.Error())
	}
}