package server

import (
	"github.com/Xhofe/alist/model"
	"github.com/gofiber/fiber/v2"
	"strconv"
)

func SaveSettings(ctx *fiber.Ctx) error {
	var req []model.SettingItem
	if err := ctx.BodyParser(&req); err != nil {
		return ErrorResp(ctx,err,400)
	}
	if err := validate.Struct(req); err != nil {
		return ErrorResp(ctx,err,400)
	}
	if err := model.SaveSettings(req); err != nil {
		return ctx.JSON(Resp{
			Code: 500,
			Msg:  err.Error(),
			Data: nil,
		})
	} else {
		return SuccessResp(ctx)
	}
}

func GetSettingsByType(ctx *fiber.Ctx) error {
	t, err := strconv.Atoi(ctx.Query("type"))
	if err != nil {
		return ErrorResp(ctx, err, 400)
	}
	settings, err := model.GetSettingByType(t)
	if err != nil {
		return ErrorResp(ctx, err, 400)
	}
	return SuccessResp(ctx,settings)
}

func GetSettingsPublic(ctx *fiber.Ctx) error {
	settings, err := model.GetSettingByType(0)
	if err != nil {
		return ErrorResp(ctx, err, 400)
	}
	return SuccessResp(ctx,settings)
}