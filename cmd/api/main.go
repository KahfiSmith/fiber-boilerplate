package main

import (
	"fmt"
	"log"
	
	"fiber-boilerplate/src/config"
	"fiber-boilerplate/src/database"
	"fiber-boilerplate/src/common/server"
	"fiber-boilerplate/src/common/middleware"
	"fiber-boilerplate/src/common/redis"
	"fiber-boilerplate/src/common/jwt"
	healthControllerPkg "fiber-boilerplate/src/modules/health/controller"
	healthServicePkg "fiber-boilerplate/src/modules/health/service"
	"fiber-boilerplate/src/modules/auth"
	"fiber-boilerplate/src/modules/auth/types"

	"github.com/gofiber/fiber/v3"
	"fiber-boilerplate/src/common/response"
)

func main() {
	// 1. load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// 2. init infrastructure (db, redis, logger)
	database.Connect(cfg.DB)
	redis.Connect(cfg.Redis)
	database.DB.AutoMigrate(&types.User{}) // simple auto-migrate

	// 3. init app
	app := fiber.New(fiber.Config{
		ErrorHandler: response.GlobalErrorHandler,
	})
	app.Use(middleware.Logger())

	// 4. init modules (manual di)
	healthService := healthServicePkg.NewHealthService(cfg.App.Name)
	healthController := healthControllerPkg.NewHealthController(healthService)

	tokenService := jwt.NewTokenService(cfg.Auth)
	authRepo := auth.NewAuthRepository()
	authService := auth.NewAuthService(authRepo, tokenService, cfg.Auth)
	authController := auth.NewAuthController(authService, cfg.Auth)

	// 5. register routes
	server.RegisterRoutes(app, server.Dependencies{
		HealthController: healthController,
		AuthController:   authController,
		Config:           cfg,
	})

	// 6. start server
	addr := fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port)
	fmt.Printf("Server starting on %s\n", addr)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}
