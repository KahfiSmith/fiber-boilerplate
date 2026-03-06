package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	config "fiber-boilerplate/pkg/configs"
	controller "fiber-boilerplate/pkg/controllers"
	repository "fiber-boilerplate/pkg/repositories"
	"fiber-boilerplate/pkg/server"
	"fiber-boilerplate/pkg/services"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Errorf("load config: %w", err))
	}

	log, err := config.NewLogger(cfg)
	if err != nil {
		panic(fmt.Errorf("init logger: %w", err))
	}
	defer log.Sync()

	db, err := config.NewGormDB(cfg, log)
	if err != nil {
		log.Fatal("failed to connect database", config.Err(err))
	}
	defer config.CloseGormDB(db)

	healthRepo := repository.NewHealthRepository(cfg.App.Name)
	healthService := services.NewHealthService(healthRepo)
	healthController := controller.NewHealthController(healthService)

	validate := config.NewValidator()
	app, err := server.New(cfg, log, db, validate, server.Dependencies{
		HealthController: healthController,
	})
	if err != nil {
		log.Fatal("failed to initialize server", config.Err(err))
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := server.Run(ctx, app, cfg, log); err != nil {
		log.Fatal("server exited with error", config.Err(err))
	}
}
