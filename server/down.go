package server

import "github.com/gofiber/fiber/v2"

func Down(ctx *fiber.Ctx) error {
	rawPath := ctx.Params("*")
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
