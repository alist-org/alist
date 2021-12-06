package server

import (
	"fmt"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"strings"
)

type Resp struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func ParsePath(rawPath string) (*model.Account, string, base.Driver, error) {
	var path, name string
	switch model.AccountsCount() {
	case 0:
		return nil, "", nil, fmt.Errorf("no accounts,please add one first")
	case 1:
		path = rawPath
		break
	default:
		paths := strings.Split(rawPath, "/")
		path = "/" + strings.Join(paths[2:], "/")
		name = paths[1]
	}
	account, ok := model.GetAccount(name)
	if !ok {
		return nil, "", nil, fmt.Errorf("no [%s] account", name)
	}
	driver, ok := base.GetDriver(account.Type)
	if !ok {
		return nil, "", nil, fmt.Errorf("no [%s] driver", account.Type)
	}
	return &account, path, driver, nil
}

func ErrorResp(c *gin.Context, err error, code int) {
	log.Error(err.Error())
	c.JSON(200, Resp{
		Code:    code,
		Message: err.Error(),
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
