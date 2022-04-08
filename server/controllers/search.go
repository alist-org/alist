package controllers

import (
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/server/common"
	"github.com/gin-gonic/gin"
)

type SearchReq struct {
	Path    string `json:"path"`
	Keyword string `json:"keyword"`
}

func Search(c *gin.Context) {
	if !conf.GetBool("enable search") {
		common.ErrorStrResp(c, "Not allowed search", 403)
		return
	}
	var req SearchReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	files, err := model.SearchByNameAndPath(req.Path, req.Keyword)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c, files)
}
