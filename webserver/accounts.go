package webserver

import (
	"github.com/byvko-dev/am-core/logs"
	"github.com/byvko-dev/am-stats-updates/scheduler"
	"github.com/gofiber/fiber/v2"
)

func updateRealmAccounts(c *fiber.Ctx) error {
	realm := c.Params("realm")
	if realm == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Realm is required"})
	}

	err := scheduler.CreateRealmTasks(scheduler.TaskTypeAccountUpdate, realm, 3, 100) // 3 tries per task
	if err != nil {
		logs.Error("Error creating snapshot tasks for realm: %s", realm)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.SendStatus(fiber.StatusAccepted)
}
