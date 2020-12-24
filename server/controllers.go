package server

import (
	"github.com/Xhofe/alist/alidrive"
	"github.com/Xhofe/alist/conf"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func Info(c *gin.Context) {
	c.JSON(200,dataResponse(conf.Conf.Info))
}

func Get(c *gin.Context) {
	var get alidrive.GetReq
	if err := c.ShouldBindJSON(&get); err != nil {
		c.JSON(200,metaResponse(400,"Bad Request"))
		return
	}
	log.Debugf("get:%v",get)
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
	c.JSON(200,dataResponse(file))
}

func List(c *gin.Context) {
	var list ListReq
	if err := c.ShouldBindJSON(&list);err!=nil {
		c.JSON(200,metaResponse(400,"Bad Request"))
		return
	}
	log.Debugf("list:%v",list)
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
		c.JSON(200,metaResponse(500,err.Error()))
		return
	}
	password:=alidrive.HasPassword(files)
	if password!="" && password!=list.Password {
		if list.Password=="" {
			c.JSON(200,metaResponse(401,"need password."))
			return
		}
		c.JSON(200,metaResponse(401,"wrong password."))
		return
	}
	paths,err:=alidrive.GetPaths(list.ParentFileId)
	if err!=nil {
		c.JSON(200,metaResponse(500,err.Error()))
		return
	}
	files.Paths=*paths
	files.Readme=alidrive.HasReadme(files)
	c.JSON(200,dataResponse(files))
}

func Search(c *gin.Context) {
	if !conf.Conf.Server.Search {
		c.JSON(200,metaResponse(403,"Not allow search."))
		return
	}
	var search alidrive.SearchReq
	if err := c.ShouldBindJSON(&search); err != nil {
		c.JSON(200,metaResponse(400,"Bad Request"))
		return
	}
	log.Debugf("search:%v",search)
	if search.Limit == 0 {
		search.Limit=50
	}
	// Search只支持0-100
	//if conf.Conf.AliDrive.MaxFilesCount!=0 {
	//	search.Limit=conf.Conf.AliDrive.MaxFilesCount
	//}
	files,err:=alidrive.Search(search.Query,search.Limit,search.OrderBy)
	if err != nil {
		c.JSON(200,metaResponse(500,err.Error()))
		return
	}
	c.JSON(200,dataResponse(files))
}