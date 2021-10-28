package server

import (
	"github.com/Xhofe/alist/utils"
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
	"net/url"
)

func Down(ctx *fiber.Ctx) error {
	rawPath, err:= url.QueryUnescape(ctx.Params("*"))
	if err != nil {
		return ErrorResp(ctx,err,500)
	}
	rawPath = utils.ParsePath(rawPath)
	log.Debugf("down: %s",rawPath)
	account, path, driver, err := ParsePath(rawPath)
	if err != nil {
		return ErrorResp(ctx, err, 500)
	}
	link, err := driver.Link(path, account)
	if err != nil {
		return ErrorResp(ctx, err, 500)
	}
	if account.Type == "native" {
		return ctx.SendFile(link)
	} else {
		return ctx.Redirect(link, 302)
	}
}
