package middlewares

import (
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/server/common"
	"github.com/Xhofe/alist/utils"
	"github.com/gin-gonic/gin"
)

func DownCheck(c *gin.Context) {
	sign := c.Query("sign")
	rawPath := c.Param("path")
	rawPath = utils.ParsePath(rawPath)
	name := utils.Base(rawPath)
	if sign == utils.SignWithToken(name, conf.Token) {
		c.Set("sign", true)
		c.Next()
		return
	}
	pw := c.Query("pw")
	if !common.CheckDownLink(utils.Dir(rawPath), pw, utils.Base(rawPath)) {
		common.ErrorResp(c, fmt.Errorf("wrong password"), 401)
		c.Abort()
		return
	}
	c.Next()
}
