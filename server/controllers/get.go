package controllers

import (
	"github.com/Xhofe/alist/alidrive"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"strings"
)

// 因为下载地址有时效，所以去掉了文件请求和直链的缓存

func Get(c *gin.Context) {
	var get alidrive.GetReq
	if err := c.ShouldBindJSON(&get); err != nil {
		c.JSON(200,metaResponse(400,"Bad Request"))
		return
	}
	log.Debugf("get:%+v",get)
	// cache
	//cacheKey:=fmt.Sprintf("%s-%s","g",get.FileId)
	//if conf.Conf.Cache.Enable {
	//	file,exist:=conf.Cache.Get(cacheKey)
	//	if exist {
	//		log.Debugf("使用了缓存:%s",cacheKey)
	//		c.JSON(200,dataResponse(file))
	//		return
	//	}
	//}
	file,err:=alidrive.GetFile(get.FileId)
	if err !=nil {
		c.JSON(200,metaResponse(500,err.Error()))
		return
	}
	paths,err:=alidrive.GetPaths(get.FileId)
	if err!=nil {
		c.JSON(200,metaResponse(500,err.Error()))
		return
	}
	file.Paths=*paths
	//if conf.Conf.Cache.Enable {
	//	conf.Cache.Set(cacheKey,file,cache.DefaultExpiration)
	//}
	c.JSON(200,dataResponse(file))
}

func Down(c *gin.Context) {
	fileIdParam:=c.Param("file_id")
	log.Debugf("down:%s",fileIdParam)
	fileId:=strings.Split(fileIdParam,"/")[1]
	//cacheKey:=fmt.Sprintf("%s-%s","d",fileId)
	//if conf.Conf.Cache.Enable {
	//	downloadUrl,exist:=conf.Cache.Get(cacheKey)
	//	if exist {
	//		log.Debugf("使用了缓存:%s",cacheKey)
	//		c.Redirect(301,downloadUrl.(string))
	//		return
	//	}
	//}
	file,err:=alidrive.GetFile(fileId)
	if err != nil {
		c.JSON(200, metaResponse(500,err.Error()))
		return
	}
	//if conf.Conf.Cache.Enable {
	//	conf.Cache.Set(cacheKey,file.DownloadUrl,cache.DefaultExpiration)
	//}
	c.Redirect(301,file.DownloadUrl)
	return
}