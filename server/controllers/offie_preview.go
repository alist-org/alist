package controllers

import (
	"github.com/Xhofe/alist/alidrive"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// handle office_preview request
func OfficePreview(c *gin.Context) {
	var req alidrive.OfficePreviewUrlReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, MetaResponse(400,"Bad Request"))
		return
	}
	log.Debugf("preview_req:%+v",req)
	preview,err:=alidrive.GetOfficePreviewUrl(req.FileId)
	if err!=nil {
		c.JSON(200, MetaResponse(500,err.Error()))
		return
	}
	c.JSON(200, DataResponse(preview))
}