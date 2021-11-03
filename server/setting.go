package server

import (
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/model"
	"github.com/gofiber/fiber/v2"
	"strings"
)

func SaveSettings(ctx *fiber.Ctx) error {
	var req []model.SettingItem
	if err := ctx.BodyParser(&req); err != nil {
		return ErrorResp(ctx, err, 400)
	}
	//if err := validate.Struct(req); err != nil {
	//	return ErrorResp(ctx, err, 400)
	//}
	if err := model.SaveSettings(req); err != nil {
		return ErrorResp(ctx, err, 500)
	} else {
		textTypes, err := model.GetSettingByKey("text types")
		if err==nil{
			conf.TextTypes = strings.Split(textTypes.Value,",")
		}
		return SuccessResp(ctx)
	}
}

func GetSettings(ctx *fiber.Ctx) error {
	settings, err := model.GetSettings()
	if err != nil {
		return ErrorResp(ctx, err, 400)
	}
	return SuccessResp(ctx, settings)
}

func GetSettingsPublic(ctx *fiber.Ctx) error {
	settings, err := model.GetSettingsPublic()
	if err != nil {
		return ErrorResp(ctx, err, 400)
	}
	return SuccessResp(ctx, settings)
}
