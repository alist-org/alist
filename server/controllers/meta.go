package controllers

import (
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/server/common"
	"github.com/Xhofe/alist/utils"
	"github.com/gin-gonic/gin"
	"strconv"
)

func GetMetas(c *gin.Context)  {
	metas,err := model.GetMetas()
	if err != nil {
		common.ErrorResp(c,err,500)
		return
	}
	common.SuccessResp(c, metas)
}

func CreateMeta(c *gin.Context)  {
	var req model.Meta
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	req.Path = utils.ParsePath(req.Path)
	if err := model.CreateMeta(req); err != nil {
		common.ErrorResp(c, err, 500)
	} else {
		common.SuccessResp(c)
	}
}

func SaveMeta(c *gin.Context)  {
	var req model.Meta
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	req.Path = utils.ParsePath(req.Path)
	if err := model.SaveMeta(req); err != nil {
		common.ErrorResp(c, err, 500)
	} else {
		common.SuccessResp(c)
	}
}

func DeleteMeta(c *gin.Context) {
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	//path = utils.ParsePath(path)
	if err := model.DeleteMeta(uint(id)); err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c)
}
