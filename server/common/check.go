package common

import (
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func Login(c *gin.Context) {
	SuccessResp(c)
}

func CheckParent(path string, password string) bool {
	meta, err := model.GetMetaByPath(path)
	if err == nil {
		if meta.Password != "" && meta.Password != password {
			return false
		}
		return true
	} else {
		if path == "/" {
			return true
		}
		return CheckParent(utils.Dir(path), password)
	}
}

func CheckDownLink(path string, passwordMd5 string, name string) bool {
	if !conf.GetBool("check down link") {
		return true
	}
	meta, err := model.GetMetaByPath(path)
	log.Debugf("check down path: %s", path)
	if err == nil {
		log.Debugf("check down link: %s,%s", meta.Password, passwordMd5)
		if meta.Password != "" && utils.SignWithPassword(name, meta.Password) != passwordMd5 {
			return false
		}
		return true
	} else {
		if !conf.GetBool("check parent folder") {
			return true
		}
		if path == "/" {
			return true
		}
		return CheckDownLink(utils.Dir(path), passwordMd5, name)
	}
}
