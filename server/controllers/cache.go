package controllers

import (
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/server/common"
	"github.com/gin-gonic/gin"
)

func ClearCache(c *gin.Context) {
	err := conf.Cache.Clear(conf.Ctx)
	if err != nil {
		common.ErrorResp(c, err, 500)
	} else {
		common.SuccessResp(c)
	}
}
