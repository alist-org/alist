package middlewares

import (
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/server/common"
	"github.com/gin-gonic/gin"
)

func AuthAdmin(c *gin.Context) {
	token := c.GetHeader("Authorization")
	userClaims, err := common.ParseToken(token)
	if err != nil {
		common.ErrorResp(c, err, 401)
		c.Abort()
		return
	}
	user, err := db.GetUserByName(userClaims.Username)
	if err != nil {
		common.ErrorResp(c, err, 401)
		c.Abort()
		return
	}
	c.Set("user", user)
	c.Next()
}
