package main

import (
	"fmt"
	"log"
	
	"fiber-boilerplate/src/common/config"
	"fiber-boilerplate/src/common/server"
	"fiber-boilerplate/src/modules/health"

	"github.com/gofiber/fiber/v3"
)

func main() {
	// 1. Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// 2. Init App
	app := fiber.New()

	// 3. Init Modules
	healthService := health.NewHealthService(cfg.App.Name)
	healthController := health.NewHealthController(healthService)

	// 4. Register Routes
	server.RegisterRoutes(app, server.Dependencies{
		HealthController: healthController,
	})

	// 5. Start server
	addr := fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port)
	fmt.Printf("Server starting on %s\n", addr)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}
