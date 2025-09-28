package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// setupVersionCommand создает команду отображения версии
func setupVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Показать информацию о версии",
		Long: `Отображение детальной информации о версии клиента GophKeeper.

Версия включает:
- Номер версии (семантическое версионирование)
- Хэш коммита Git
- Дата и время сборки`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			displayVersionInfo()
		},
	}
}

// displayVersionInfo отображает информацию о версии
func displayVersionInfo() {
	fmt.Printf("GophKeeper Client\n")
	fmt.Printf("Версия: %s\n", Version)
	fmt.Printf("Коммит: %s\n", Commit)
	fmt.Printf("Собрано: %s\n", Date)

	// Дополнительная информация о возможностях
	fmt.Printf("\nВозможности:\n")
	fmt.Printf("  • Хранение паролей, карт, текстов и файлов\n")
	fmt.Printf("  • Сквозное шифрование данных\n")
	fmt.Printf("  • Синхронизация между устройствами\n")
	fmt.Printf("  • Работа в оффлайн-режиме\n")

	fmt.Printf("\nСервер: %s\n", server)
	fmt.Printf("Данные: %s\n", dataDir)

	if cfgFile != "" {
		fmt.Printf("Конфиг: %s\n", cfgFile)
	} else {
		fmt.Printf("Конфиг: используется значение по умолчанию\n")
	}
}
