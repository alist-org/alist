package middlewares

import (
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/server/common"
	"github.com/gin-gonic/gin"
	"sync"
	"time"
)

var m sync.Map

type Client struct {
	count int
	t     *time.Timer
}

var (
	defaultDuration = time.Minute * 5
	defaultTimes    = 8
)

func Auth(c *gin.Context) {
	ip := c.ClientIP()
	if v, ok := m.Load(ip); ok {
		client := v.(Client)
		if client.count > defaultTimes {
			common.ErrorStrResp(c, "Too many unsuccessful sign-in attempts have been made using an incorrect password. Try again later.", 403)
			client.t.Reset(defaultDuration)
			return
		}
	}
	token := c.GetHeader("Authorization")
	//password, err := model.GetSettingByKey("password")
	//if err != nil {
	//	if err == gorm.ErrRecordNotFound {
	//		common.ErrorResp(c, fmt.Errorf("password not set"), 400)
	//		return
	//	}
	//	common.ErrorResp(c, err, 500)
	//	return
	//}
	//if token != utils.GetMD5Encode(password.Value) {
	if token != conf.Token {
		common.ErrorStrResp(c, "Invalid token", 401)
		v, ok := m.Load(ip)
		var client Client
		if ok {
			client = v.(Client)
		} else {
			client = Client{
				count: 0,
				t: time.AfterFunc(defaultDuration, func() {
					m.Delete(ip)
				}),
			}
		}
		client.count++
		client.t.Reset(defaultDuration)
		m.Store(ip, client)
		return
	}
	c.Next()
}
