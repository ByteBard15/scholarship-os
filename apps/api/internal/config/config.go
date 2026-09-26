package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Environment    string
	Port           string
	WebOrigin      string
	UploadDir      string
	MaxUploadBytes int64
	Database       DatabaseConfig
	Auth           AuthConfig
}

type AuthConfig struct {
	SystemAPIKey string
	SessionTTL   time.Duration
	TokenPepper  string
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
	maxOpen, err := integer("DATABASE_MAX_OPEN_CONNS", 10)
	if err != nil {
		return Config{}, err
	}
	maxIdle, err := integer("DATABASE_MAX_IDLE_CONNS", 5)
	if err != nil {
		return Config{}, err
	}
	maxUploadBytes, err := positiveInt64("MAX_UPLOAD_BYTES", 10*1024*1024)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		Environment:    env,
		Port:           value("PORT", "8080"),
		WebOrigin:      value("WEB_ORIGIN", "http://localhost:5173"),
		UploadDir:      value("UPLOAD_DIR", "../../.var/uploads"),
		MaxUploadBytes: maxUploadBytes,
		Auth:           AuthConfig{SystemAPIKey: os.Getenv("SYSTEM_API_KEY"), SessionTTL: 24 * time.Hour, TokenPepper: os.Getenv("AUTH_TOKEN_HASH_PEPPER")},
		Database: DatabaseConfig{
			Host: value("DATABASE_HOST", "localhost"), Port: dbPort,
			Name: value("DATABASE_NAME", "scholarship_os"), User: value("DATABASE_USER", "postgres"),
			Password: value("DATABASE_PASSWORD", "postgres"), SSLMode: value("DATABASE_SSLMODE", "disable"),
			MaxOpenConns: maxOpen, MaxIdleConns: maxIdle, ConnMaxLifetime: time.Hour,
		},
	}
	if raw := os.Getenv("USER_SESSION_TTL"); raw != "" {
		ttl, parseErr := time.ParseDuration(raw)
		if parseErr != nil || ttl <= 0 {
			return Config{}, fmt.Errorf("USER_SESSION_TTL must be a positive duration")
		}
		cfg.Auth.SessionTTL = ttl
	}
	if cfg.Auth.SystemAPIKey != "" && (!strings.HasPrefix(cfg.Auth.SystemAPIKey, "sys_") || len(cfg.Auth.SystemAPIKey) < 36) {
		return Config{}, fmt.Errorf("SYSTEM_API_KEY must use the sys_ prefix and contain at least 32 characters")
	}
	if env == "production" {
		for key, current := range map[string]string{"DATABASE_HOST": os.Getenv("DATABASE_HOST"), "DATABASE_NAME": os.Getenv("DATABASE_NAME"), "DATABASE_USER": os.Getenv("DATABASE_USER"), "DATABASE_PASSWORD": os.Getenv("DATABASE_PASSWORD"), "WEB_ORIGIN": os.Getenv("WEB_ORIGIN"), "UPLOAD_DIR": os.Getenv("UPLOAD_DIR"), "SYSTEM_API_KEY": os.Getenv("SYSTEM_API_KEY")} {
			if current == "" {
				return Config{}, fmt.Errorf("%s is required in production", key)
			}
		}
	}
	return cfg, nil
}

func positiveInt64(key string, fallback int64) (int64, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || v <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", key)
	}
	return v, nil
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
