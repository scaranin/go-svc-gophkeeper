package config

import (
	"fmt"
	"time"
)

// ClientConfig содержит настройки CLI клиента
type ClientConfig struct {
	ServerURL      string        `yaml:"server_url" env:"CLIENT_SERVER_URL" env-default:"localhost:50051"`
	Timeout        time.Duration `yaml:"timeout" env:"CLIENT_TIMEOUT" env-default:"30s"`
	ConfigPath     string        `yaml:"-"`
	DataDir        string        `yaml:"data_dir" env:"CLIENT_DATA_DIR" env-default:"~/.gophkeeper"`
	MasterPassword string        `yaml:"-"`
}

// Validate проверяет корректность клиентской конфигурации
func (c *ClientConfig) Validate() error {
	if c.ServerURL == "" {
		return fmt.Errorf("server URL is required")
	}
	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	return nil
}

// LoadClientConfig загружает конфигурацию клиента
func LoadClientConfig(configPath string) (*ClientConfig, error) {
	var cfg ClientConfig

	return &cfg, nil
}
