package controllers

import (
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	common2 "github.com/alist-org/alist/v3/server/common"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"strconv"
)

func ListMetas(c *gin.Context) {
	var req common2.PageReq
	if err := c.ShouldBind(&req); err != nil {
		common2.ErrorResp(c, err, 400)
		return
	}
	log.Debugf("%+v", req)
	metas, total, err := db.GetMetas(req.PageIndex, req.PageSize)
	if err != nil {
		common2.ErrorResp(c, err, 500, true)
		return
	}
	common2.SuccessResp(c, common2.PageResp{
		Content: metas,
		Total:   total,
	})
}

func CreateMeta(c *gin.Context) {
	var req model.Meta
	if err := c.ShouldBind(&req); err != nil {
		common2.ErrorResp(c, err, 400)
		return
	}
	req.Path = utils.StandardizePath(req.Path)
	if err := db.CreateMeta(&req); err != nil {
		common2.ErrorResp(c, err, 500)
	} else {
		common2.SuccessResp(c)
	}
}

func UpdateMeta(c *gin.Context) {
	var req model.Meta
	if err := c.ShouldBind(&req); err != nil {
		common2.ErrorResp(c, err, 400)
		return
	}
	req.Path = utils.StandardizePath(req.Path)
	if err := db.UpdateMeta(&req); err != nil {
		common2.ErrorResp(c, err, 500)
	} else {
		common2.SuccessResp(c)
	}
}

func DeleteMeta(c *gin.Context) {
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		common2.ErrorResp(c, err, 400)
		return
	}
	if err := db.DeleteMetaById(uint(id)); err != nil {
		common2.ErrorResp(c, err, 500)
		return
	}
	common2.SuccessResp(c)
}
