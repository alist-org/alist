package server

import (
	"fmt"
	"github.com/Xhofe/alist/drivers"
	"github.com/Xhofe/alist/model"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"strings"
)

var validate = validator.New()

type Resp struct {
	Code int `json:"code"`
	Msg string `json:"msg"`
	Data interface{} `json:"data"`
}

func ParsePath(rawPath string) (*model.Account,string,drivers.Driver,error) {
	var path,name string
	switch model.AccountsCount() {
	case 0:
		return nil,"",nil,fmt.Errorf("no accounts,please add one first")
	case 1:
		path = rawPath
		break
	default:
		paths := strings.Split(rawPath,"/")
		path = strings.Join(paths[2:],"/")
		name = paths[1]
	}
	account,ok := model.GetAccount(name)
	if !ok {
		return nil,"",nil,fmt.Errorf("no [%s] account", name)
	}
	driver,ok := drivers.GetDriver(account.Type)
	if !ok {
		return nil,"",nil,fmt.Errorf("no [%s] driver",account.Type)
	}
	return &account,path,driver,nil
}

func ErrorResp(ctx *fiber.Ctx,err error,code int) error {
	return ctx.JSON(Resp{
		Code: code,
		Msg:  err.Error(),
		Data: nil,
	})
}

func SuccessResp(ctx *fiber.Ctx, data ...interface{}) error {
	if len(data) == 0 {
		return ctx.JSON(Resp{
			Code: 200,
			Msg:  "success",
			Data: nil,
		})
	}
	return ctx.JSON(Resp{
		Code: 200,
		Msg:  "success",
		Data: data[0],
	})
}