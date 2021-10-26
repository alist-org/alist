package server

import "github.com/gofiber/fiber/v2"

func InitApiRouter(app *fiber.App) {

	app.Get("/d/*", Down)

	public := app.Group("/api/public")
	{
		// TODO check accounts
		public.Post("/path", Path)
	}

	admin := app.Group("/api/admin")
	{
		// TODO auth
		admin.Get("/settings", GetSettingsByType)
		admin.Post("/settings", SaveSettings)
		admin.Post("/account", SaveAccount)
		admin.Get("/accounts", GetAccounts)
		admin.Delete("/account", DeleteAccount)
		admin.Get("/drivers", GetDrivers)
	}
}
