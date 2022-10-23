package webserver

import (
	"github.com/byvko-dev/am-stats-updates/internal/scheduler"
	api "github.com/byvko-dev/am-types/api/generic/v1"
	"github.com/gofiber/fiber/v2"
)

func genericGlossaryHandler(handler func(string, int) error, taskType string, tries int) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var response api.ResponseWithError
		err := handler(taskType, tries)
		if err != nil {
			response.Error.Context = err.Error()
			response.Error.Message = "Error creating tasks"
			return c.Status(500).JSON(response)
		}
		return c.Status(fiber.StatusAccepted).JSON(response)
	}
}

var updateVehicles = genericGlossaryHandler(scheduler.CreateGlossaryTasks, scheduler.TaskTypeUpdateGlossaryVehicles, 3)
var updateAverages = genericGlossaryHandler(scheduler.CreateGlossaryTasks, scheduler.TaskTypeUpdateGlossaryAverages, 3)
var updateAchievements = genericGlossaryHandler(scheduler.CreateGlossaryTasks, scheduler.TaskTypeUpdateGlossaryAchievements, 3)
