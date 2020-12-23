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