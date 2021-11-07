package server

import (
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	log "github.com/sirupsen/logrus"
	"net/url"
	"path/filepath"
)

func Down(ctx *fiber.Ctx) error {
	rawPath, err:= url.PathUnescape(ctx.Params("*"))
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
	if account.Type == "Native" {
		return ctx.SendFile(link)
	} else {
		return ctx.Redirect(link, 302)
	}
}

func Proxy(ctx *fiber.Ctx) error {
	rawPath, err:= url.PathUnescape(ctx.Params("*"))
	if err != nil {
		return ErrorResp(ctx,err,500)
	}
	rawPath = utils.ParsePath(rawPath)
	log.Debugf("proxy: %s",rawPath)
	account, path, driver, err := ParsePath(rawPath)
	if err != nil {
		return ErrorResp(ctx, err, 500)
	}
	if !account.Proxy && utils.GetFileType(filepath.Ext(rawPath))!=conf.TEXT {
		return ErrorResp(ctx,fmt.Errorf("[%s] not allowed proxy",account.Name),403)
	}
	link, err := driver.Link(path, account)
	if err != nil {
		return ErrorResp(ctx, err, 500)
	}
	if account.Type == "Native" {
		return ctx.SendFile(link)
	} else {
		driver.Proxy(ctx)
		if err := proxy.Do(ctx, link); err != nil {
			log.Errorf("proxy error: %s", err)
			return ErrorResp(ctx,err,500)
		}
		// Remove Server header from response
		ctx.Response().Header.Del(fiber.HeaderServer)
		ctx.Set("Access-Control-Allow-Origin","*")
		log.Debugf("proxy hedaer: %+v", ctx.Response().Header.String())
		return nil
	}
}