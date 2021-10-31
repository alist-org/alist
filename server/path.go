package server

import (
	"fmt"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
)

type PathReq struct {
	Path     string `json:"Path"`
	Password string `json:"Password"`
}

func Path(ctx *fiber.Ctx) error {
	var req PathReq
	if err := ctx.BodyParser(&req); err != nil {
		return ErrorResp(ctx, err, 400)
	}
	req.Path = utils.ParsePath(req.Path)
	log.Debugf("path: %s",req.Path)
	meta, err := model.GetMetaByPath(req.Path)
	if err == nil {
		if meta.Password != "" && meta.Password!= req.Password {
			return ErrorResp(ctx,fmt.Errorf("wrong password"),401)
		}
		// TODO hide or ignore?
	}
	if model.AccountsCount() > 1 && req.Path == "/" {
		return ctx.JSON(Resp{
			Code: 200,
			Message:  "folder",
			Data: model.GetAccountFiles(),
		})
	}
	account, path, driver, err := ParsePath(req.Path)
	if err != nil {
		return ErrorResp(ctx, err, 500)
	}
	file, files, err := driver.Path(path, account)
	if err != nil {
		return ErrorResp(ctx, err, 500)
	}
	if file != nil {
		return ctx.JSON(Resp{
			Code: 200,
			Message:  "file",
			Data: []*model.File{file},
		})
	} else {
		return ctx.JSON(Resp{
			Code: 200,
			Message:  "folder",
			Data: files,
		})
	}
}

func Link(ctx *fiber.Ctx) error {
	var req PathReq
	if err := ctx.BodyParser(&req); err != nil {
		return ErrorResp(ctx, err, 400)
	}
	rawPath := req.Path
	rawPath = utils.ParsePath(rawPath)
	log.Debugf("link: %s",rawPath)
	account, path, driver, err := ParsePath(rawPath)
	if err != nil {
		return ErrorResp(ctx, err, 500)
	}
	link, err := driver.Link(path, account)
	if err != nil {
		return ErrorResp(ctx, err, 500)
	}
	if account.Type == "Native" {
		return SuccessResp(ctx, fiber.Map{
			"url":"",
		})
	} else {
		return SuccessResp(ctx,fiber.Map{
			"url":link,
		})
	}
}

func Preview(ctx *fiber.Ctx) error {
	var req PathReq
	if err := ctx.BodyParser(&req); err != nil {
		return ErrorResp(ctx, err, 400)
	}
	rawPath := req.Path
	rawPath = utils.ParsePath(rawPath)
	log.Debugf("preview: %s",rawPath)
	account, path, driver, err := ParsePath(rawPath)
	if err != nil {
		return ErrorResp(ctx, err, 500)
	}
	data, err := driver.Preview(path, account)
	if err != nil {
		return ErrorResp(ctx,err,500)
	}else {
		return SuccessResp(ctx,data)
	}
}