package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func InitApiRouter(app *fiber.App) {

	// TODO from settings
	app.Use(cors.New())
	app.Get("/d/*", Down)
	// TODO check allow proxy?
	app.Get("/p/*", Proxy)

	api := app.Group("/api")
	api.Use(SetSuccess)
	public := api.Group("/public")
	{
		public.Post("/path", CheckAccount, Path)
		public.Post("/preview", CheckAccount, Preview)
		public.Get("/settings", GetSettingsPublic)
		public.Post("/link", CheckAccount, Link)
	}

	admin := api.Group("/admin")
	{
		admin.Use(Auth)
		admin.Get("/login", Login)
		admin.Get("/settings", GetSettings)
		admin.Post("/settings", SaveSettings)
		admin.Post("/account", SaveAccount)
		admin.Get("/accounts", GetAccounts)
		admin.Delete("/account", DeleteAccount)
		admin.Get("/drivers", GetDrivers)
		admin.Get("/clear_cache",ClearCache)
	}
}
