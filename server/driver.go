package server

import (
	"github.com/Xhofe/alist/drivers"
	"github.com/gofiber/fiber/v2"
)

func GetDrivers(ctx *fiber.Ctx) error {
	return SuccessResp(ctx, drivers.GetDrivers())
}