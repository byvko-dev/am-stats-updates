package webserver

import (
	"github.com/byvko-dev/am-stats-updates/internal/scheduler"
	"github.com/gofiber/fiber/v2"
)

func genericGlossaryHandler(handler func(string, int) error, taskType string, tries int) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		err := handler(taskType, tries)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.SendStatus(fiber.StatusAccepted)
	}
}

var updateVehicles = genericGlossaryHandler(scheduler.CreateGlossaryTasks, scheduler.TaskTypeUpdateGlossaryVehicles, 3)
var updateAverages = genericGlossaryHandler(scheduler.CreateGlossaryTasks, scheduler.TaskTypeUpdateGlossaryAverages, 3)
var updateAchievements = genericGlossaryHandler(scheduler.CreateGlossaryTasks, scheduler.TaskTypeUpdateGlossaryAchievements, 3)
