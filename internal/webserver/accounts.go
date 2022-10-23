package webserver

import (
	"github.com/byvko-dev/am-core/logs"
	"github.com/byvko-dev/am-stats-updates/internal/scheduler"
	api "github.com/byvko-dev/am-types/api/generic/v1"
	"github.com/gofiber/fiber/v2"
)

func updateRealmAccounts(c *fiber.Ctx) error {
	var response api.ResponseWithError
	realm := c.Params("realm")
	if realm == "" {
		response.Error.Message = "Realm is required"
		response.Error.Context = "params.realm is empty"
		return c.Status(fiber.StatusBadRequest).JSON(response)
	}

	err := scheduler.CreateRealmTasks(scheduler.TaskTypeAccountUpdate, realm, 3, 100) // 3 tries per task
	if err != nil {
		logs.Error("Error creating snapshot tasks for realm: %s", realm)
		response.Error.Context = err.Error()
		response.Error.Message = "Error creating snapshot tasks"
		return c.Status(fiber.StatusInternalServerError).JSON(response)
	}

	return c.Status(fiber.StatusAccepted).JSON(response)
}
