package base

import (
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	log "github.com/sirupsen/logrus"
)

func KeyCache(path string, account *model.Account) string {
	path = utils.ParsePath(path)
	return fmt.Sprintf("%s%s", account.Name, path)
}

func SetCache(path string, obj interface{}, account *model.Account) error {
	return conf.Cache.Set(conf.Ctx, KeyCache(path, account), obj, nil)
}

func GetCache(path string, account *model.Account) (interface{}, error) {
	return conf.Cache.Get(conf.Ctx, KeyCache(path, account))
}

func DeleteCache(path string, account *model.Account) error {
	err := conf.Cache.Delete(conf.Ctx, KeyCache(path, account))
	log.Debugf("delete cache %s: %+v", path, err)
	return err
}
