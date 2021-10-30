package server

import (
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/model"
	"github.com/gofiber/fiber/v2"
	"strconv"
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

func GetSettingsByGroup(ctx *fiber.Ctx) error {
	t, err := strconv.Atoi(ctx.Query("type"))
	if err != nil {
		return ErrorResp(ctx, err, 400)
	}
	settings, err := model.GetSettingsByGroup(t)
	if err != nil {
		return ErrorResp(ctx, err, 400)
	}
	return SuccessResp(ctx, settings)
}

func GetSettings(ctx *fiber.Ctx) error {
	settings, err := model.GetSettings()
	if err != nil {
		return ErrorResp(ctx, err, 400)
	}
	return SuccessResp(ctx, settings)
}

func GetSettingsPublic(ctx *fiber.Ctx) error {
	settings, err := model.GetSettingsByGroup(0)
	if err != nil {
		return ErrorResp(ctx, err, 400)
	}
	return SuccessResp(ctx, settings)
}
