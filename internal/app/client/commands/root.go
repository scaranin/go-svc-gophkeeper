package commands

import (
	"fmt"

	"go-svc-gophkeeper/internal/app/client"
	"go-svc-gophkeeper/internal/config"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
	server  string
	dataDir string

	Version string
	Commit  string
	Date    string

	app *client.App
)

// SetupRootCommand создает и настраивает корневую команду
func SetupRootCommand(version, commit, date string) *cobra.Command {
	Version = version
	Commit = commit
	Date = date

	rootCmd := &cobra.Command{
		Use:   "gophkeeper",
		Short: "Безопасный менеджер паролей с синхронизацией",
		Long: `GophKeeper - безопасный менеджер паролей с клиент-серверной синхронизацией.

Возможности:
- Хранение паролей, банковских карт, текстовых заметок и бинарных данных
- Сквозное шифрование
- Синхронизация между устройствами
- Работа в оффлайн-режиме`,
		Version:               fmt.Sprintf("%s (%s) собрано %s", Version, Commit, Date),
		SilenceUsage:          true,
		SilenceErrors:         true,
		DisableFlagsInUseLine: true,
		PersistentPreRunE:     setupApp,
		PersistentPostRun:     cleanupApp,
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "/.congigs/client.yaml", "файл конфигурации")
	rootCmd.PersistentFlags().StringVarP(&server, "server", "s", "localhost:50051", "адрес сервера")
	rootCmd.PersistentFlags().StringVarP(&dataDir, "data-dir", "d", "~/.go-svc-gophkeeper", "директория для локального хранения данных")

	rootCmd.AddCommand(
		setupAuthCommand(),
		setupSecretCommand(),
		setupVersionCommand(),
	)

	return rootCmd
}

// setupApp инициализирует приложение
func setupApp(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadClientConfig(cfgFile)
	if err != nil {
		return fmt.Errorf("ошибка загрузки конфигурации: %w", err)
	}

	if server != "" {
		cfg.ServerURL = server
	}
	if dataDir != "" {
		cfg.DataDir = dataDir
	}

	app, err = client.NewApp(cfg)
	if err != nil {
		return fmt.Errorf("ошибка создания приложения: %w", err)
	}

	return nil
}

// cleanupApp освобождает ресурсы
func cleanupApp(cmd *cobra.Command, args []string) {
	if app != nil {
		app.Close()
	}
}
