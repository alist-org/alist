package server

import (
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/gin-gonic/gin"
)

func GetMetas(c *gin.Context)  {
	metas,err := model.GetMetas()
	if err != nil {
		ErrorResp(c,err,500)
		return
	}
	SuccessResp(c, metas)
}

func CreateMeta(c *gin.Context)  {
	var req model.Meta
	if err := c.ShouldBind(&req); err != nil {
		ErrorResp(c, err, 400)
		return
	}
	req.Path = utils.ParsePath(req.Path)
	if err := model.CreateMeta(req); err != nil {
		ErrorResp(c, err, 500)
	} else {
		SuccessResp(c)
	}
}

func SaveMeta(c *gin.Context)  {
	var req model.Meta
	if err := c.ShouldBind(&req); err != nil {
		ErrorResp(c, err, 400)
		return
	}
	req.Path = utils.ParsePath(req.Path)
	if err := model.SaveMeta(req); err != nil {
		ErrorResp(c, err, 500)
	} else {
		SuccessResp(c)
	}
}

func DeleteMeta(c *gin.Context) {
	path := c.Query("path")
	//path = utils.ParsePath(path)
	if err := model.DeleteMeta(path); err != nil {
		ErrorResp(c, err, 500)
		return
	}
	SuccessResp(c)
}
