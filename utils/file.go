package utils

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
)

func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func CreatNestedFile(path string) (*os.File, error) {
	basePath := filepath.Dir(path)
	if !Exists(basePath) {
		err := os.MkdirAll(basePath, 0700)
		if err != nil {
			log.Errorf("无法创建目录，%s", err)
			return nil, err
		}
	}
	return os.Create(path)
}

func WriteToYml(src string,conf interface{}){
	data,err := yaml.Marshal(conf)
	if err!=nil {
		log.Errorf("Conf转[]byte失败:%s",err.Error())
	}
	err = ioutil.WriteFile(src,data,0777)
	if err!=nil {
		log.Errorf("写yml文件失败",err.Error())
	}
}