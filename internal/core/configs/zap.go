package config

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(cfg Config) (*zap.Logger, error) {
	level := zapcore.InfoLevel
	if parsed, err := zapcore.ParseLevel(cfg.Log.Level); err == nil {
		level = parsed
	}

	zapCfg := zap.NewDevelopmentConfig()
	if cfg.App.Env == "production" {
		zapCfg = zap.NewProductionConfig()
	}

	if cfg.Log.Encoding != "" {
		zapCfg.Encoding = cfg.Log.Encoding
	}
	zapCfg.Level = zap.NewAtomicLevelAt(level)

	return zapCfg.Build()
}

func Err(err error) zap.Field {
	return zap.Error(err)
}
