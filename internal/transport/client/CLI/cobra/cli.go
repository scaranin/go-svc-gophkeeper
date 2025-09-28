package cobra

import (
	"go-svc-gophkeeper/internal/app/client/commands"
)

// RunCLI запускает CLI приложение через Cobra
func RunCLI(version, commit, date string) {
	// Инициализация команд
	rootCmd := commands.SetupRootCommand(version, commit, date)

	// Запуск Cobra
	if err := rootCmd.Execute(); err != nil {
		// Cobra сам обработает ошибки и покажет help
	}
}
