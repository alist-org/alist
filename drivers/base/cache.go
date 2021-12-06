package base

import (
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
)

func KeyCache(path string, account *model.Account) string {
	path = utils.ParsePath(path)
	return fmt.Sprintf("%s%s", account.Name, path)
}

func SetCache(path string, obj interface{}, account *model.Account) error {
	return conf.Cache.Set(conf.Ctx, KeyCache(path, account), obj, nil)
}

func GetCache(path string, account *model.Account) (interface{}, error) {
	return conf.Cache.Get(conf.Ctx, fmt.Sprintf("%s%s", KeyCache(path, account)))
}

func DeleteCache(path string, account *model.Account) error {
	return conf.Cache.Delete(conf.Ctx, fmt.Sprintf("%s%s", KeyCache(path, account)))
}
