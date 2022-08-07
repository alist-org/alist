package utils

import (
	"io/ioutil"

	json "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
)

var Json = json.ConfigCompatibleWithStandardLibrary

// WriteJsonToFile write struct to json file
func WriteJsonToFile(src string, conf interface{}) bool {
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
