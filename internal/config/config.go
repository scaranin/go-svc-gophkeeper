package config

import (
	"time"
)

// Config главная конфигурационная структура
type Config struct {
	Server     ServerConfig     `yaml:"server"`
	Database   DatabaseConfig   `yaml:"database"`
	Auth       AuthConfig       `yaml:"auth"`
	Encryption EncryptionConfig `yaml:"encryption"`
}

// ServerConfig содержит настройки gRPC сервера
type ServerConfig struct {
	Host            string        `yaml:"host" env:"SERVER_HOST" env-default:"0.0.0.0"`
	Port            int           `yaml:"port" env:"SERVER_PORT" env-default:"50051"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env:"SERVER_SHUTDOWN_TIMEOUT" env-default:"30s"`
}

// DatabaseConfig содержит настройки подключения к БД
type DatabaseConfig struct {
	DSN            string `yaml:"dsn" env:"DATABASE_DSN" env-default:"postgres://user:pass@localhost:5432/gophkeeper?sslmode=disable"`
	MigrationsPath string `yaml:"migrations_path" env:"DATABASE_MIGRATIONS_PATH" env-default:"./migrations/postgres"`
}

// AuthConfig содержит настройки аутентификации
type AuthConfig struct {
	JWTSecret         string        `yaml:"jwt_secret" env:"AUTH_JWT_SECRET" env-default:"TZOY_ZHIV"`
	AccessTokenExpiry time.Duration `yaml:"access_token_expiry" env:"AUTH_ACCESS_TOKEN_EXPIRY" env-default:"15m"`
}

// EncryptionConfig содержит настройки шифрования
type EncryptionConfig struct {
	Key string `yaml:"key" env:"ENCRYPTION_KEY" env-default:"lf989jkj9f09dfis0dkf09i0fw9kf0sk"`
}
