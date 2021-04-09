package bootstrap

import (
	"crypto/tls"
	"github.com/Xhofe/alist/conf"
	log "github.com/sirupsen/logrus"
	"net/http"
)

// init request client
func InitClient() {
	log.Infof("初始化client...")
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	conf.Client = &http.Client{Transport: tr}
}
