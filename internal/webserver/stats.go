package webserver

import (
	"strings"

	"github.com/byvko-dev/am-core/logs"
	"github.com/byvko-dev/am-stats-updates/internal/scheduler"
	"github.com/byvko-dev/am-stats-updates/internal/workers/snapshots"
	"github.com/gofiber/fiber/v2"
)

func recordRealmSessions(c *fiber.Ctx) error {
	realm := c.Params("realm")
	if realm == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Realm is required"})
	}

	err := scheduler.CreateRealmTasks(scheduler.TaskTypeSnapshot, realm, 3, 50) // 3 tries per task
	if err != nil {
		logs.Error("Error creating snapshot tasks for realm: %s", realm)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.SendStatus(fiber.StatusAccepted)
}

func recordPlayerSessions(c *fiber.Ctx) error {
	ids := c.Query("ids")
	realm := c.Params("realm")
	manual := c.Query("manual") == "true"
	if realm == "" || ids == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Realm and playerIDs are required"})
	}

	result, _, err := snapshots.SaveAccountSnapshots(realm, strings.Split(ids, ","), manual)
	if err != nil {
		logs.Error("Error recording snapshots for players: %s", ids)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(result)
}
