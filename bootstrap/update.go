package bootstrap

import (
	"encoding/json"
	"github.com/Xhofe/alist/conf"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

type GithubRelease struct {
	TagName	string `json:"tag_name"`
	HtmlUrl	string `json:"html_url"`
	Body	string `json:"body"`
}

func CheckUpdate() {
	url:="https://api.github.com/repos/Xhofe/alist/releases/latest"
	resp,err:=http.Get(url)
	if err!=nil {
		log.Warnf("检查更新失败:%s",err.Error())
		return
	}
	body,err:=ioutil.ReadAll(resp.Body)
	if err!=nil {
		log.Warnf("读取更新内容失败:%s",err.Error())
		return
	}
	var release GithubRelease
	err = json.Unmarshal(body,&release)
	if err!=nil {
		log.Warnf("解析更新失败:%s",err.Error())
		return
	}
	if conf.VERSION == release.TagName {
		log.Infof("当前已是最新版本:%s",release.TagName)
	}else {
		log.Infof("发现新版本:%s",release.TagName)
		log.Infof("请至'%s'获取更新.",release.HtmlUrl)
	}
}