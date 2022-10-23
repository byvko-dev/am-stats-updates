package webserver

import (
	"strings"

	"github.com/byvko-dev/am-core/logs"
	"github.com/byvko-dev/am-stats-updates/internal/scheduler"
	"github.com/byvko-dev/am-stats-updates/internal/workers/snapshots"
	api "github.com/byvko-dev/am-types/api/generic/v1"
	"github.com/gofiber/fiber/v2"
)

func recordRealmSessions(c *fiber.Ctx) error {
	var response api.ResponseWithError
	realm := c.Params("realm")
	if realm == "" {
		response.Error.Message = "Realm is required"
		response.Error.Context = "params.realm is empty"
		return c.Status(fiber.StatusBadRequest).JSON(response)
	}

	err := scheduler.CreateRealmTasks(scheduler.TaskTypeSnapshot, realm, 3, 50) // 3 tries per task
	if err != nil {
		logs.Error("Error creating snapshot tasks for realm: %s", realm)
		response.Error.Context = err.Error()
		response.Error.Message = "Error creating snapshot tasks"
		return c.Status(fiber.StatusInternalServerError).JSON(response)
	}

	return c.Status(fiber.StatusAccepted).JSON(response)
}

func recordPlayerSessions(c *fiber.Ctx) error {
	var response api.ResponseWithError
	ids := c.Query("ids")
	realm := c.Params("realm")
	manual := c.Query("manual") == "true"
	if realm == "" || ids == "" {
		response.Error.Context = "params.realm or query.ids is empty"
		response.Error.Message = "Realm and ids are required"
		return c.Status(fiber.StatusBadRequest).JSON(response)
	}

	result, _, err := snapshots.SaveAccountSnapshots(realm, strings.Split(ids, ","), manual)
	if err != nil {
		logs.Error("Error recording snapshots for players: %s", ids)
		response.Error.Context = err.Error()
		response.Error.Message = "Error recording snapshots"
		return c.Status(fiber.StatusInternalServerError).JSON(response)
	}

	response.Data = result
	return c.JSON(response)
}
