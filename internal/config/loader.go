package config

import (
	"os"
	"path/filepath"
	"strings"
)

// Load загружает конфигурацию из файла и переменных окружения
func Load(configPath string) (*Config, error) {
	var cfg Config

	return &cfg, nil
}

// findConfigFile ищет конфигурационный файл в стандартных местах
func findConfigFile() string {
	possiblePaths := []string{
		"./config.yaml",
		"./config.yml",
		"./config/config.yaml",
		"./config/config.yml",
		"/etc/gophkeeper/config.yaml",
		"~/.config/gophkeeper/config.yaml",
	}

	for _, path := range possiblePaths {
		if strings.HasPrefix(path, "~/") {
			home, _ := os.UserHomeDir()
			path = filepath.Join(home, path[2:])
		}

		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}
