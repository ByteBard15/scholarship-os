package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Environment string
	Port        string
	WebOrigin   string
	Database    DatabaseConfig
}

type DatabaseConfig struct {
	Host            string
	Port            int
	Name            string
	User            string
	Password        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func Load() (Config, error) {
	env := value("APP_ENV", "development")
	dbPort, err := integer("DATABASE_PORT", 5432)
	if err != nil {
		return Config{}, err
	}
	maxOpen, err := integer("DATABASE_MAX_OPEN_CONNS", 25)
	if err != nil {
		return Config{}, err
	}
	maxIdle, err := integer("DATABASE_MAX_IDLE_CONNS", 5)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		Environment: env,
		Port:        value("APP_PORT", "8080"),
		WebOrigin:   value("WEB_ORIGIN", "http://localhost:5173"),
		Database: DatabaseConfig{
			Host: value("DATABASE_HOST", "localhost"), Port: dbPort,
			Name: value("DATABASE_NAME", "scholarship_os"), User: value("DATABASE_USER", "postgres"),
			Password: value("DATABASE_PASSWORD", "postgres"), SSLMode: value("DATABASE_SSLMODE", "disable"),
			MaxOpenConns: maxOpen, MaxIdleConns: maxIdle, ConnMaxLifetime: time.Hour,
		},
	}
	if env == "production" {
		for key, current := range map[string]string{"DATABASE_HOST": os.Getenv("DATABASE_HOST"), "DATABASE_NAME": os.Getenv("DATABASE_NAME"), "DATABASE_USER": os.Getenv("DATABASE_USER"), "DATABASE_PASSWORD": os.Getenv("DATABASE_PASSWORD"), "WEB_ORIGIN": os.Getenv("WEB_ORIGIN")} {
			if current == "" {
				return Config{}, fmt.Errorf("%s is required in production", key)
			}
		}
	}
	return cfg, nil
}

func (c DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC", c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode)
}

func value(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
func integer(key string, fallback int) (int, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}
	return v, nil
}
