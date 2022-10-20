package webserver

import (
	"github.com/byvko-dev/am-core/helpers/env"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func Start() {
	app := fiber.New()
	app.Use(logger.New())

	v1 := app.Group("/v1")

	glossary := v1.Group("/glossary")
	glossary.Get("/vehicles", updateVehicles)
	glossary.Get("/averages", updateAverages)
	glossary.Get("/achievements", updateAchievements)

	realm := v1.Group("/realm/:realm")
	realm.Get("/accounts", updateRealmAccounts)
	realm.Get("/sessions/all", recordRealmSessions)
	realm.Get("/sessions/players", recordPlayerSessions)

	panic(app.Listen(":" + env.MustGetString("PORT")))
}
