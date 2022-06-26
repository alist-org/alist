package common

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type Resp struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func ErrorResp(c *gin.Context, err error, code int, l ...bool) {
	if len(l) != 0 && l[0] {
		log.Errorf("%+v", err)
	}
	c.JSON(200, Resp{
		Code:    code,
		Message: err.Error(),
		Data:    nil,
	})
	c.Abort()
}

func ErrorStrResp(c *gin.Context, str string, code int) {
	log.Error(str)
	c.JSON(200, Resp{
		Code:    code,
		Message: str,
		Data:    nil,
	})
	c.Abort()
}

func SuccessResp(c *gin.Context, data ...interface{}) {
	if len(data) == 0 {
		c.JSON(200, Resp{
			Code:    200,
			Message: "success",
			Data:    nil,
		})
		return
	}
	c.JSON(200, Resp{
		Code:    200,
		Message: "success",
		Data:    data[0],
	})
}
