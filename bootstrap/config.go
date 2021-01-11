package bootstrap

import (
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/utils"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strings"
)

func ReadConf(config string) bool {
	log.Infof("读取配置文件...")
	if !utils.Exists(config) {
		log.Infof("找不到配置文件:%s",config)
		return false
	}
	confFile,err:=ioutil.ReadFile(config)
	if err !=nil {
		log.Errorf("读取配置文件时发生错误:%s",err.Error())
		return false
	}
	err = yaml.Unmarshal(confFile, conf.Conf)
	if err !=nil {
		log.Errorf("加载配置文件时发生错误:%s",err.Error())
		return false
	}
	log.Debugf("config:%+v",conf.Conf)
	conf.Origins = strings.Split(conf.Conf.Server.SiteUrl,",")
	return true
}