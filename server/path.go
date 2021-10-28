package server

import (
	"github.com/Xhofe/alist/model"
	"github.com/gofiber/fiber/v2"
	"strings"
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
	if !strings.HasPrefix(req.Path, "/") {
		req.Path = "/"+req.Path
	}
	if model.AccountsCount() > 1 && req.Path == "/" {
		return ctx.JSON(Resp{
			Code: 200,
			Msg:  "folder",
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
			Msg:  "file",
			Data: []*model.File{file},
		})
	} else {
		return ctx.JSON(Resp{
			Code: 200,
			Msg:  "folder",
			Data: files,
		})
	}
}
