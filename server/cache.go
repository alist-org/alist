package server

import (
	"github.com/Xhofe/alist/conf"
	"github.com/gofiber/fiber/v2"
)

func ClearCache(ctx *fiber.Ctx) error {
	err := conf.Cache.Clear(conf.Ctx)
	if err != nil {
		return ErrorResp(ctx,err,500)
	}else {
		return SuccessResp(ctx)
	}
}