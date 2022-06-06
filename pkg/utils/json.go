package utils

import (
	json "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
)

var Json = json.ConfigCompatibleWithStandardLibrary

// WriteToJson write struct to json file
func WriteToJson(src string, conf interface{}) bool {
	data, err := Json.MarshalIndent(conf, "", "  ")
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
