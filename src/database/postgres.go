package database

import (
	"fiber-boilerplate/src/config"
	"log/slog"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect(cfg config.DBConfig) {
	var err error

	DB, err = gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		slog.Error("db_connect_failed", slog.Any("error", err))
		return
	}

	sqlDB, err := DB.DB()
	if err != nil {
		slog.Error("db_sql_handle_failed", slog.Any("error", err))
		return
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	slog.Info("database_connected")
}
