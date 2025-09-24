package config

import (
	"time"
)

// Config главная конфигурационная структура
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Auth     AuthConfig     `yaml:"auth"`
}

// ServerConfig содержит настройки gRPC сервера
type ServerConfig struct {
	Host            string        `yaml:"host" env:"SERVER_HOST" env-default:"0.0.0.0"`
	Port            int           `yaml:"port" env:"SERVER_PORT" env-default:"50051"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env:"SERVER_SHUTDOWN_TIMEOUT" env-default:"30s"`
}

// DatabaseConfig содержит настройки подключения к БД
type DatabaseConfig struct {
	DSN string `yaml:"dsn" env:"DATABASE_DSN" env-default:"postgres://user:pass@localhost:5432/gophkeeper?sslmode=disable"`
}

// AuthConfig содержит настройки аутентификации
type AuthConfig struct {
	JWTSecret         string        `yaml:"jwt_secret" env:"AUTH_JWT_SECRET" env-default:"your-super-secret-jwt-key"`
	AccessTokenExpiry time.Duration `yaml:"access_token_expiry" env:"AUTH_ACCESS_TOKEN_EXPIRY" env-default:"15m"`
}
