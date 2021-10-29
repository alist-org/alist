package server

import (
	"fmt"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func Auth(ctx *fiber.Ctx) error {
	token := ctx.Get("Authorization")
	password, err := model.GetSettingByKey("password")
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrorResp(ctx, fmt.Errorf("password not set"), 400)
		}
		return ErrorResp(ctx, err, 500)
	}
	if token != utils.GetMD5Encode(password.Value) {
		return ErrorResp(ctx, fmt.Errorf("wrong password"), 401)
	}
	return ctx.Next()
}

func Login(ctx *fiber.Ctx) error {
	return SuccessResp(ctx)
}

func SetSuccess(ctx *fiber.Ctx) error {
	ctx.Status(200)
	return ctx.Next()
}

func CheckAccount(ctx *fiber.Ctx) error {
	if model.AccountsCount() == 0 {
		return ErrorResp(ctx,fmt.Errorf("no accounts,please add one first"),1001)
	}
	return ctx.Next()
}