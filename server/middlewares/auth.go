package middlewares

import (
	"github.com/Xhofe/alist/server/common"
	"github.com/Xhofe/alist/utils"
	"github.com/gin-gonic/gin"
)

func Auth(c *gin.Context) {
	accessToken := c.GetHeader("Authorization")
	if !utils.VerifyAccessToken(accessToken) {
		common.ErrorStrResp(c, "Invalid token", 401)
		return
	}
	c.Next()
}
