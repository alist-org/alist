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

// list request bean
type ListReq struct {
	Password	string	`json:"password"`
	alidrive.ListReq
}

// handle list request
func List(c *gin.Context) {
	var list ListReq
	if err := c.ShouldBindJSON(&list);err!=nil {
		c.JSON(200, controllers.MetaResponse(400,"Bad Request"))
		return
	}
	log.Debugf("list:%+v",list)
	// cache
	cacheKey:=fmt.Sprintf("%s-%s-%s","l",list.ParentFileId,list.Password)
	if conf.Conf.Cache.Enable {
		files,exist:=conf.Cache.Get(cacheKey)
		if exist {
			log.Debugf("使用了缓存:%s",cacheKey)
			c.JSON(200, controllers.DataResponse(files))
			return
		}
	}
	var (
		files *alidrive.Files
		err error
	)
	if list.Limit == 0 {
		list.Limit=50
	}
	if conf.Conf.AliDrive.MaxFilesCount!=0 {
		list.Limit=conf.Conf.AliDrive.MaxFilesCount
	}
	if list.ParentFileId == "root" {
		files,err=alidrive.GetRoot(list.Limit,list.Marker,list.OrderBy,list.OrderDirection)
	}else {
		files,err=alidrive.GetList(list.ParentFileId,list.Limit,list.Marker,list.OrderBy,list.OrderDirection)
	}
	if err!=nil {
		c.JSON(200, controllers.MetaResponse(500,err.Error()))
		return
	}
	password:=alidrive.HasPassword(files)
	if password!="" && password!=list.Password {
		if list.Password=="" {
			c.JSON(200, controllers.MetaResponse(401,"need password."))
			return
		}
		c.JSON(200, controllers.MetaResponse(401,"wrong password."))
		return
	}
	paths,err:=alidrive.GetPaths(list.ParentFileId)
	if err!=nil {
		c.JSON(200, controllers.MetaResponse(500,err.Error()))
		return
	}
	files.Paths=*paths
	//files.Readme=alidrive.HasReadme(files)
	if conf.Conf.Cache.Enable {
		conf.Cache.Set(cacheKey,files,cache.DefaultExpiration)
	}
	c.JSON(200, controllers.DataResponse(files))
}