package server

import "github.com/gofiber/fiber/v2"

func InitApiRouter(app *fiber.App) {

	app.Get("/d/*", Down)

	public := app.Group("/api/public")
	{
		// TODO check accounts
		public.Post("/path", Path)
		public.Get("/settings", GetSettingsPublic)
	}

	admin := app.Group("/api/admin")
	{
		admin.Use(Auth)
		admin.Get("/settings", GetSettingsByType)
		admin.Post("/settings", SaveSettings)
		admin.Post("/account", SaveAccount)
		admin.Get("/accounts", GetAccounts)
		admin.Delete("/account", DeleteAccount)
		admin.Get("/drivers", GetDrivers)
	}
}
