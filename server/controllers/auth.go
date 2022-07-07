package controllers

import (
	"time"

	"github.com/Xhofe/go-cache"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/gin-gonic/gin"
)

var loginCache = cache.NewMemCache[int]()
var (
	defaultDuration = time.Minute * 5
	defaultTimes    = 5
)

type LoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Login(c *gin.Context) {
	// check count of login
	ip := c.ClientIP()
	count, ok := loginCache.Get(ip)
	if ok && count >= defaultTimes {
		common.ErrorStrResp(c, "Too many unsuccessful sign-in attempts have been made using an incorrect username or password. Try again later.", 403)
		loginCache.Expire(ip, defaultDuration)
		return
	}
	// check username
	var req LoginReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	user, err := db.GetUserByName(req.Username)
	if err != nil {
		common.ErrorResp(c, err, 400)
		loginCache.Set(ip, count+1)
		return
	}
	// validate password
	if err := user.ValidatePassword(req.Password); err != nil {
		common.ErrorResp(c, err, 400)
		loginCache.Set(ip, count+1)
		return
	}
	// generate token
	token, err := common.GenerateToken(user.Username)
	if err != nil {
		common.ErrorResp(c, err, 400, true)
		return
	}
	common.SuccessResp(c, gin.H{"token": token})
	loginCache.Del(ip)
}

// CurrentUser get current user by token
// if token is empty, return guest user
func CurrentUser(c *gin.Context) {
	user := c.MustGet("user").(*model.User)
	user.Password = ""
	common.SuccessResp(c, gin.H{"user": user})
}
