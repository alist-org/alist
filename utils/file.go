package utils

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Exists determine whether the file exists
func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// CreatNestedFile create nested file
func CreatNestedFile(path string) (*os.File, error) {
	basePath := filepath.Dir(path)
	if !Exists(basePath) {
		err := os.MkdirAll(basePath, 0700)
		if err != nil {
			log.Errorf("can't create foler，%s", err)
			return nil, err
		}
	}
	return os.Create(path)
}

// WriteToJson write struct to json file
func WriteToJson(src string, conf interface{}) bool {
	data, err := json.MarshalIndent(conf,"","  ")
	if err != nil {
		log.Errorf("failed convert Conf to []byte:%s", err.Error())
		return false
	}
	err = ioutil.WriteFile(src, data, 0777)
	if err != nil {
		log.Errorf("failed to write json file:%s", err.Error())
		return false
	}
	return true
}