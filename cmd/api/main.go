package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"fiber-boilerplate/src/common/jwt"
	"fiber-boilerplate/src/common/middleware"
	"fiber-boilerplate/src/common/redis"
	"fiber-boilerplate/src/common/response"
	"fiber-boilerplate/src/common/server"
	"fiber-boilerplate/src/config"
	"fiber-boilerplate/src/database"
	"fiber-boilerplate/src/modules/auth"
	"fiber-boilerplate/src/modules/auth/types"
	healthControllerPkg "fiber-boilerplate/src/modules/health/controller"
	healthServicePkg "fiber-boilerplate/src/modules/health/service"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	database.Connect(cfg.DB)
	redis.Connect(cfg.Redis)
	database.DB.AutoMigrate(&types.User{})

	app := fiber.New(fiber.Config{
		ErrorHandler: response.GlobalErrorHandler,
	})

	app.Use(recover.New())
	app.Use(helmet.New())

	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.Auth.FrontendOrigin},
		AllowCredentials: true,
		AllowHeaders:     []string{"Authorization", "Content-Type", "X-CSRF-Token"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
	}))

	app.Use(middleware.Logger())

	healthService := healthServicePkg.NewHealthService(cfg.App.Name)
	healthController := healthControllerPkg.NewHealthController(healthService)

	tokenService := jwt.NewTokenService(cfg.Auth)
	authRepo := auth.NewAuthRepository()
	refreshRepo := auth.NewRefreshRepository()
	authService := auth.NewAuthService(authRepo, refreshRepo, tokenService, cfg.Auth)
	oauthService := auth.NewOAuthService(cfg.OAuth, cfg.Auth, authRepo, refreshRepo, tokenService)
	authController := auth.NewAuthController(authService, oauthService, cfg.Auth)

	server.RegisterRoutes(app, server.Dependencies{
		HealthController: healthController,
		AuthController:   authController,
		Config:           cfg,
	})

	go func() {
		addr := fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port)
		fmt.Printf("Server starting on %s\n", addr)
		if err := app.Listen(addr); err != nil {
			log.Fatalf("server exited: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	fmt.Println("\nGracefully shutting down server...")

	if err := app.Shutdown(); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	fmt.Println("Server was successful shutdown.")
}
