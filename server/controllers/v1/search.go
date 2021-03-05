package v1

import (
	"fmt"
	"github.com/Xhofe/alist/alidrive"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/server/controllers"
	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
)

// handle search request
func Search(c *gin.Context) {
	if !conf.Conf.Server.Search {
		c.JSON(200, controllers.MetaResponse(403, "Not allow search."))
		return
	}
	var search alidrive.SearchReq
	if err := c.ShouldBindJSON(&search); err != nil {
		c.JSON(200, controllers.MetaResponse(400, "Bad Request"))
		return
	}
	log.Debugf("search:%+v", search)
	// cache
	cacheKey := fmt.Sprintf("%s-%s", "s", search.Query)
	if conf.Conf.Cache.Enable {
		files, exist := conf.Cache.Get(cacheKey)
		if exist {
			log.Debugf("使用了缓存:%s", cacheKey)
			c.JSON(200, controllers.DataResponse(files))
			return
		}
	}
	if search.Limit == 0 {
		search.Limit = 50
	}
	// Search只支持0-100
	//if conf.Conf.AliDrive.MaxFilesCount!=0 {
	//	search.Limit=conf.Conf.AliDrive.MaxFilesCount
	//}
	files, err := alidrive.Search(search.Query, search.Limit, search.OrderBy)
	if err != nil {
		c.JSON(200, controllers.MetaResponse(500, err.Error()))
		return
	}
	if conf.Conf.Cache.Enable {
		conf.Cache.Set(cacheKey, files, cache.DefaultExpiration)
	}
	c.JSON(200, controllers.DataResponse(files))
}
