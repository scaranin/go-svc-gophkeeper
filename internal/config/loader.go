package config

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Load загружает конфигурацию из файла и переменных окружения
func Load(configPath string) (*Config, error) {
	var cfg Config

	if configPath == "" {
		configPath = findConfigFile()
		if configPath == "" {
			return &cfg, nil
		}
	}

	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// findConfigFile ищет конфигурационный файл в стандартных местах
func findConfigFile() string {
	possiblePaths := []string{
		"./configs/server.yaml",
		"./server.yaml",
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
