package bootstrap

import (
	"github.com/Xhofe/alist/conf"
	log "github.com/sirupsen/logrus"
	"net/http"
)

// init request client
func InitClient() {
	log.Infof("初始化client...")
	conf.Client = &http.Client{}
}
