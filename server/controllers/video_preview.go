package controllers

import (
	"github.com/Xhofe/alist/alidrive"
	"github.com/Xhofe/alist/utils"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type VideoPreviewReq struct {
	FileId string `json:"file_id" binding:"required"`
}

// handle video_preview request
func VideoPreview(c *gin.Context) {
	drive := utils.GetDriveByName(c.Param("drive"))
	if drive == nil {
		c.JSON(200, MetaResponse(400, "drive isn't exist."))
		return
	}
	var req VideoPreviewReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, MetaResponse(400, "Bad Request:"+err.Error()))
		return
	}
	log.Debugf("preview_req:%+v", req)
	preview, err := alidrive.GetVideoPreviewUrl(req.FileId, drive)
	if err != nil {
		c.JSON(200, MetaResponse(500, err.Error()))
		return
	}
	c.JSON(200, DataResponse(preview))
}
