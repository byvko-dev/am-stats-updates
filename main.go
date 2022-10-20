package main

import (
	"context"
	"os"
	"os/signal"
	"strings"

	"github.com/byvko-dev/am-core/helpers/env"
	"github.com/byvko-dev/am-core/logs"
	"github.com/byvko-dev/am-stats-updates/core/messaging"
	"github.com/byvko-dev/am-stats-updates/scheduler"
	"github.com/byvko-dev/am-stats-updates/workers"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/robfig/cron/v3"
)

func main() {
	// Task workers
	cancel := make(chan int, 1)
	go startTaskQueue(cancel)

	// Task scheduler
	stopScheduler := startScheduler()

	// Web server
	go startWebServer()

	// Wait for system signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	cancel <- 1
	stopScheduler()
	logs.Info("Shutting down...")
}

func startWebServer() {
	app := fiber.New()
	app.Use(logger.New())

	v1 := app.Group("/v1")

	v1.Get("/update-glossary", func(c *fiber.Ctx) error {
		err := scheduler.CreateGlossaryTasks(scheduler.TaskTypeUpdateGlossary, 3)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.SendStatus(fiber.StatusAccepted)
	})

	realm := v1.Group("/realm/:realm")
	// Reset all sessions on realm
	realm.Get("/reset-sessions", func(c *fiber.Ctx) error {
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
	})
	// Update all accounts on realm
	realm.Get("/update-accounts", func(c *fiber.Ctx) error {
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
	})

	panic(app.Listen(":" + env.MustGetString("PORT")))
}

func startTaskQueue(cancel chan int) {
	messaging.Connect(env.MustGetString("MESSAGING_URI"))
	err := workers.StartUpdateWorkers(cancel) // Will execute all tasks in the queue every 5 min
	if err != nil {
		panic(err)
	}
}

func startScheduler() func() context.Context {
	nameSlice := strings.Split(env.MustGetString("POD_NAME"), "-")
	withCron := nameSlice[len(nameSlice)-1] == "0"
	if withCron {
		logs.Info("Starting cron jobs")
		runner := cron.New()
		// Update players and sessions
		runner.AddFunc("0 0 * * *", createGlossaryTasks)                  // Glossary update
		runner.AddFunc("0 9 * * *", func() { createRealmTasks("NA") })    // NA
		runner.AddFunc("0 1 * * *", func() { createRealmTasks("EU") })    // EU
		runner.AddFunc("0 18 * * *", func() { createRealmTasks("ASIA") }) // ASIA
		runner.Start()
		return runner.Stop
	}
	return func() context.Context { return context.Background() }
}

func createGlossaryTasks() {
	err := scheduler.CreateGlossaryTasks(scheduler.TaskTypeUpdateGlossary, 3)
	if err != nil {
		logs.Error("Error creating glossary tasks")
	}
}

func createRealmTasks(realm string) {
	err := scheduler.CreateRealmTasks(scheduler.TaskTypeAccountUpdate, realm, 3, 100) // 3 tries per task
	if err != nil {
		logs.Error("Error creating account update tasks for realm: %s", realm)
	}
	err = scheduler.CreateRealmTasks(scheduler.TaskTypeSnapshot, realm, 3, 50) // 3 tries per task
	if err != nil {
		logs.Error("Error creating snapshot tasks for realm: %s", realm)
	}
}
