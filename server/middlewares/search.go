package middlewares

import (
	"github.com/alist-org/alist/v3/internal2/conf"
	"github.com/alist-org/alist/v3/internal2/errs"
	"github.com/alist-org/alist/v3/internal2/setting"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/gin-gonic/gin"
)

func SearchIndex(c *gin.Context) {
	mode := setting.GetStr(conf.SearchIndex)
	if mode == "none" {
		common.ErrorResp(c, errs.SearchNotAvailable, 500)
		c.Abort()
	} else {
		c.Next()
	}
}
