package middlewares

import (
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/server/common"
	"github.com/gin-gonic/gin"
)

func Auth(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token != conf.Token {
		common.ErrorStrResp(c, "Invalid token", 401)
		return
	}
	c.Next()
}
