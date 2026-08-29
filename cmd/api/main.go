package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"fiber-boilerplate/src/common/jwt"
	"fiber-boilerplate/src/common/logger"
	"fiber-boilerplate/src/common/mail"
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
	"github.com/gofiber/fiber/v3/middleware/requestid"
)

func main() {
	log := logger.New()
	slog.SetDefault(log)

	cfg, err := config.Load()
	if err != nil {
		log.Error("failed to load config", slog.Any("error", err))
		os.Exit(1)
	}

	database.Connect(cfg.DB)
	redis.Connect(cfg.Redis)
	database.DB.AutoMigrate(&types.User{})

	app := fiber.New(fiber.Config{
		ErrorHandler: response.GlobalErrorHandler,
	})

	app.Use(requestid.New())
	app.Use(recover.New())
	app.Use(helmet.New())

	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.Auth.FrontendOrigin},
		AllowCredentials: true,
		AllowHeaders:     []string{"Authorization", "Content-Type", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
	}))

	app.Use(middleware.Logger(log))

	healthService := healthServicePkg.NewHealthService(cfg.App.Name)
	healthController := healthControllerPkg.NewHealthController(healthService)

	tokenService := jwt.NewTokenService(cfg.Auth)
	authRepo := auth.NewAuthRepository()
	refreshRepo := auth.NewRefreshRepository()
	logMailer := mail.NewLogMailer(cfg.App.Name)
	authService := auth.NewAuthService(authRepo, refreshRepo, tokenService, logMailer, cfg.Auth)
	oauthService := auth.NewOAuthService(cfg.OAuth, cfg.Auth, authRepo, refreshRepo, tokenService)
	authController := auth.NewAuthController(authService, oauthService, cfg.Auth)

	server.RegisterRoutes(app, server.Dependencies{
		HealthController: healthController,
		AuthController:   authController,
		Config:           cfg,
	})

	go func() {
		addr := fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port)
		log.Info("server_starting", slog.String("addr", addr))
		if err := app.Listen(addr); err != nil {
			log.Error("server_exited", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Info("graceful_shutdown_started")

	if err := app.Shutdown(); err != nil {
		log.Error("server_forced_shutdown", slog.Any("error", err))
		os.Exit(1)
	}

	log.Info("server_shutdown_complete")
}
