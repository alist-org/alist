package middlewares

import (
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/model"
	common2 "github.com/alist-org/alist/v3/server/common"
	"github.com/gin-gonic/gin"
)

// Auth is a middleware that checks if the user is logged in.
// if token is empty, set user to guest
func Auth(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		guest, err := db.GetGuest()
		if err != nil {
			common2.ErrorResp(c, err, 500)
			c.Abort()
			return
		}
		c.Set("user", guest)
		c.Next()
		return
	}
	userClaims, err := common2.ParseToken(token)
	if err != nil {
		common2.ErrorResp(c, err, 401, true)
		c.Abort()
		return
	}
	user, err := db.GetUserByName(userClaims.Username)
	if err != nil {
		common2.ErrorResp(c, err, 401)
		c.Abort()
		return
	}
	c.Set("user", user)
	c.Next()
}

func AuthAdmin(c *gin.Context) {
	user := c.MustGet("user").(*model.User)
	if !user.IsAdmin() {
		common2.ErrorStrResp(c, "You are not an admin", 403, true)
		c.Abort()
	} else {
		c.Next()
	}
}
