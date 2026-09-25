package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/example/scholarship-os/apps/api/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Open(ctx context.Context, cfg config.DatabaseConfig, appEnv string, log *slog.Logger) (*gorm.DB, error) {
	level := logger.Warn
	if appEnv == "development" {
		level = logger.Info
	}
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{Logger: logger.Default.LogMode(level), NowFunc: func() time.Time { return time.Now().UTC() }})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("access database pool: %w", err)
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}
	log.Info("database connection established")
	return db, nil
}
