package main

import (
	"fmt"
	"log"
	
	"fiber-boilerplate/src/common/config"
	"fiber-boilerplate/src/common/database"
	"fiber-boilerplate/src/common/server"
	"fiber-boilerplate/src/modules/health"
	"fiber-boilerplate/src/modules/auth"

	"github.com/gofiber/fiber/v3"
)

func main() {
	// 1. Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// 2. Init Infrastructure (DB, Redis, Logger)
	database.Connect(cfg.DB)
	database.DB.AutoMigrate(&auth.User{}) // Simple auto-migrate

	// 3. Init App
	app := fiber.New()

	// 4. Init Modules (Manual DI)
	healthService := health.NewHealthService(cfg.App.Name)
	healthController := health.NewHealthController(healthService)

	authRepo := auth.NewAuthRepository()
	authService := auth.NewAuthService(authRepo)
	authController := auth.NewAuthController(authService)

	// 5. Register Routes
	server.RegisterRoutes(app, server.Dependencies{
		HealthController: healthController,
		AuthController:   authController,
	})

	// 6. Start server
	addr := fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port)
	fmt.Printf("Server starting on %s\n", addr)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}
