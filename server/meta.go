package server

import (
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/gofiber/fiber/v2"
)

func GetMetas(ctx *fiber.Ctx) error {
	metas,err := model.GetMetas()
	if err != nil {
		return ErrorResp(ctx,err,500)
	}
	return SuccessResp(ctx, metas)
}

func SaveMeta(ctx *fiber.Ctx) error {
	var req model.Meta
	if err := ctx.BodyParser(&req); err != nil {
		return ErrorResp(ctx, err, 400)
	}
	if err := validate.Struct(req); err != nil {
		return ErrorResp(ctx, err, 400)
	}
	req.Path = utils.ParsePath(req.Path)
	if err := model.SaveMeta(req); err != nil {
		return ErrorResp(ctx, err, 500)
	} else {
		return SuccessResp(ctx)
	}
}

func DeleteMeta(ctx *fiber.Ctx) error {
	path := ctx.Query("path")
	//path = utils.ParsePath(path)
	if err := model.DeleteMeta(path); err != nil {
		return ErrorResp(ctx, err, 500)
	}
	return SuccessResp(ctx)
}
