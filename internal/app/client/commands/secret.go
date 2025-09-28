package commands

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// setupSecretCommand создает команды управления секретами
func setupSecretCommand() *cobra.Command {
	secretCmd := &cobra.Command{
		Use:   "secret",
		Short: "Управление секретами",
		Long: `Создание, просмотр, список и синхронизация секретов.

Поддерживаемые типы секретов:
- login: логин/пароль (веб-сайты, приложения)
- card: банковские карты (номер, срок, CVV)
- text: произвольные текстовые данные (заметки, коды)
- binary: файлы и бинарные данные (документы, изображения)`,
	}

	secretCmd.AddCommand(
		setupCreateCommand(),
		setupListCommand(),
		setupGetCommand(),
		setupSyncCommand(),
	)

	return secretCmd
}

// setupCreateCommand создает команду создания секрета
func setupCreateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "Создать новый секрет",
		Long: `Интерактивное создание нового секрета.

Команда проведет вас через процесс создания секрета
с выбором типа и вводом необходимых данных.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Интерактивное создание секрета:")

			ctx := context.Background()
			if err := app.CreateSecret(ctx); err != nil {
				return fmt.Errorf("ошибка создания секрета: %w", err)
			}

			fmt.Println("Секрет успешно создан")
			return nil
		},
	}
}

// setupListCommand создает команду списка секретов
func setupListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Показать все секреты",
		Long: `Отображение списка всех локально сохраненных секретов.

Используйте команду sync для обновления списка
секретов с сервера.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Список секретов:")

			ctx := context.Background()
			if err := app.ListSecrets(ctx); err != nil {
				return fmt.Errorf("ошибка получения списка секретов: %w", err)
			}

			return nil
		},
	}
}

// setupGetCommand создает команду получения секрета
func setupGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Получить детали секрета",
		Long: `Получение детальной информации о конкретном секрете.

Секрет будет расшифрован и отображен в читаемом формате.
Для просмотра требуется указать ID секрета.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, _ := cmd.Flags().GetString("id")
			fmt.Printf("Получение секрета: %s\n", id)

			ctx := context.Background()
			if err := app.GetSecret(ctx, id); err != nil {
				return fmt.Errorf("ошибка получения секрета: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().String("id", "", "ID секрета (обязательно)")
	cmd.MarkFlagRequired("id")

	return cmd
}

// setupSyncCommand создает команду синхронизации
func setupSyncCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Синхронизировать с сервером",
		Long: `Синхронизация локальных секретов с сервером.

Эта команда выполнит:
- Загрузку новых секретов с сервера
- Выгрузку локальных изменений
- Разрешение конфликтов (по умолчанию приоритет у сервера)`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Синхронизация с сервером...")

			ctx := context.Background()
			if err := app.Sync(ctx); err != nil {
				return fmt.Errorf("ошибка синхронизации: %w", err)
			}

			fmt.Println("Синхронизация завершена успешно")
			return nil
		},
	}
}
