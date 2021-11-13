package server

import (
	"github.com/Xhofe/alist/conf"
	"github.com/gin-gonic/gin"
)

func ClearCache(c *gin.Context) {
	err := conf.Cache.Clear(conf.Ctx)
	if err != nil {
		ErrorResp(c, err, 500)
	} else {
		SuccessResp(c)
	}
}
