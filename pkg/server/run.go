package server

import (
	"context"
	"fmt"

	config "fiber-boilerplate/pkg/configs"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

func Run(ctx context.Context, app *fiber.App, cfg config.Config, log *zap.Logger) error {
	errCh := make(chan error, 1)

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.App.ShutdownTimeout)
		defer cancel()

		if err := app.ShutdownWithContext(shutdownCtx); err != nil {
			log.Error("failed to shutdown server", zap.Error(err))
		}
	}()

	go func() {
		addr := fmt.Sprintf("%s:%s", cfg.App.Host, cfg.App.Port)
		log.Info("starting server", zap.String("address", addr))
		errCh <- app.Listen(addr, fiber.ListenConfig{
			EnablePrefork: cfg.Fiber.Prefork,
		})
	}()

	return <-errCh
}
