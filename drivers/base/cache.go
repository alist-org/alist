package base

import (
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	log "github.com/sirupsen/logrus"
	"strings"
)

func KeyCache(path string, account *model.Account) string {
	//path = utils.ParsePath(path)
	key := utils.ParsePath(utils.Join(account.Name, path))
	log.Debugln("cache key: ", key)
	return key
}

func SaveSearchFiles[T model.ISearchFile](key string, obj []T) {
	if strings.Contains(key, ".balance") {
		return
	}
	err := model.DeleteSearchFilesByPath(key)
	if err != nil {
		log.Errorln("failed create search files", err)
		return
	}
	files := make([]model.SearchFile, len(obj))
	for i := 0; i < len(obj); i++ {
		files[i] = model.SearchFile{
			Path: key,
			Name: obj[i].GetName(),
			Size: obj[i].GetSize(),
			Type: obj[i].GetType(),
		}
	}
	err = model.CreateSearchFiles(files)
	if err != nil {
		log.Errorln("failed create search files", err)
	}
}

func SetCache[T model.ISearchFile](path string, obj []T, account *model.Account) error {
	key := KeyCache(path, account)
	if conf.GetBool("enable search") {
		go SaveSearchFiles(key, obj)
	}
	return conf.Cache.Set(conf.Ctx, key, obj, nil)
}

func GetCache(path string, account *model.Account) (interface{}, error) {
	return conf.Cache.Get(conf.Ctx, KeyCache(path, account))
}

func DeleteCache(path string, account *model.Account) error {
	err := conf.Cache.Delete(conf.Ctx, KeyCache(path, account))
	log.Debugf("delete cache %s: %+v", path, err)
	return err
}
