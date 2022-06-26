package middlewares

import (
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/server/common"
	"github.com/gin-gonic/gin"
)

// Auth is a middleware that checks if the user is logged in.
// if token is empty, set user to guest
func Auth(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		guest, err := db.GetGuest()
		if err != nil {
			common.ErrorResp(c, err, 500, true)
			c.Abort()
			return
		}
		c.Set("user", guest)
		c.Next()
		return
	}
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
