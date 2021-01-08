package utils

import (
	log "github.com/sirupsen/logrus"
	"net/url"
)

func GetCode(rawUrl string) string {
	u,err:=url.Parse(rawUrl)
	if err!=nil {
		log.Errorf("解析url出错:%s",err.Error())
		return ""
	}
	code:=u.Query().Get("code")
	return code
}

func ContainsString(array []string, val string) (index int) {
	index = -1
	for i := 0; i < len(array); i++ {
		if array[i] == val {
			index = i
			return
		}
	}
	return
}