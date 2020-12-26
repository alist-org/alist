package bootstrap

import (
	"github.com/Xhofe/alist/conf"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func InitClient()  {
	log.Infof("初始化client...")
	conf.Client=&http.Client{}
}