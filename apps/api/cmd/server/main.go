package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/example/scholarship-os/apps/api/internal/config"
	"github.com/example/scholarship-os/apps/api/internal/database"
	"github.com/example/scholarship-os/apps/api/internal/profile"
	"github.com/example/scholarship-os/apps/api/internal/server"
	"github.com/example/scholarship-os/apps/api/internal/user"
	"github.com/go-playground/validator/v10"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(log)
	cfg, err := config.Load()
	if err != nil {
		log.Error("load configuration", "error", err)
		os.Exit(1)
	}
	startupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	db, err := database.Open(startupCtx, cfg.Database, cfg.Environment, log)
	if err != nil {
		log.Error("connect database", "error", err)
		os.Exit(1)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Error("access database pool", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			log.Error("close database", "error", err)
		}
	}()
	validate := validator.New(validator.WithRequiredStructEnabled())
	userRepo := user.NewGORMRepository(db)
	profileRepo := profile.NewGORMRepository(db)
	userService := user.NewService(userRepo)
	profileService := profile.NewService(profileRepo, profileRepo, userRepo)
	handler := server.NewRouter(log, sqlDB, cfg.WebOrigin, user.NewHandler(userService, validate, log), profile.NewHandler(profileService, validate, log))
	httpServer := &http.Server{Addr: ":" + cfg.Port, Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	serverErrors := make(chan error, 1)
	go func() {
		log.Info("api listening", "port", cfg.Port, "environment", cfg.Environment)
		serverErrors <- httpServer.ListenAndServe()
	}()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-signals:
		log.Info("shutdown signal received", "signal", sig.String())
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Error("http server failed", "error", err)
		}
	}
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Error("graceful shutdown failed", "error", err)
	}
}
