package commands

import (
	"context"
	"fmt"
	v1 "go-svc-gophkeeper/gen/go/v1"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/timestamppb"
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
	var (
		secretType string
		name       string
		desc       string
		website    string
		tags       string
		dataFile   string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Создать новый секрет",
		Long: `Создание нового секрета.

Поддерживаемые типы секретов: login, text, binary, card.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			var secretTypeEnum v1.SecretType
			switch secretType {
			case "login":
				secretTypeEnum = v1.SecretType_SECRET_TYPE_LOGIN
			case "text":
				secretTypeEnum = v1.SecretType_SECRET_TYPE_TEXT
			case "binary":
				secretTypeEnum = v1.SecretType_SECRET_TYPE_BINARY
			case "card":
				secretTypeEnum = v1.SecretType_SECRET_TYPE_CARD
			default:
				return fmt.Errorf("неподдерживаемый тип секрета: %s. Поддерживаемые: login, text, binary, card", secretType)
			}

			var encryptedData []byte
			if dataFile != "" {
				data, err := os.ReadFile(dataFile)
				if err != nil {
					return err
				}
				encryptedData = data
			}

			metadata := &v1.Metadata{
				Name:        name,
				Description: desc,
				Website:     website,
				Tags:        tags,
			}

			if err := app.CreateSecret(ctx, secretTypeEnum, name, metadata, encryptedData); err != nil {
				return err
			}

			fmt.Println("Секрет успешно создан")
			return nil
		},
	}

	cmd.Flags().StringVarP(&secretType, "type", "t", "login", "Тип секрета (login, text, binary, card)")
	cmd.Flags().StringVarP(&name, "name", "n", "", "Название секрета (обязательно)")
	cmd.Flags().StringVarP(&desc, "desc", "d", "", "Описание секрета")
	cmd.Flags().StringVarP(&website, "website", "w", "", "Веб-сайт (для логинов)")
	cmd.Flags().StringVarP(&tags, "tags", "g", "", "Теги (через запятую)")
	cmd.Flags().StringVarP(&dataFile, "file", "f", "", "Файл с зашифрованными данными")

	_ = cmd.MarkFlagRequired("name")

	return cmd
}

// setupListCommand создает команду вывода списка секретов
func setupListCommand() *cobra.Command {
	var (
		includeDeleted bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Вывести список всех секретов",
		Long: `Вывод списка всех секретов пользователя.

По умолчанию показываются только активные секреты.
Используйте флаг --deleted для отображения удаленных секретов.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			if err := app.ListSecrets(ctx, includeDeleted); err != nil {
				return err
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&includeDeleted, "deleted", "d", false, "Показать удаленные секреты")

	return cmd
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
				return err
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
	var (
		lastSyncFile string
		lastSyncTime string
	)

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Синхронизировать данные с сервером",
		Long: `Синхронизация локальных данных с сервером.

Для указания времени последней синхронизации используйте:
- --file для чтения из файла (формат RFC3339)
- --time для прямого указания времени (формат RFC3339)

Пример времени: 2023-10-01T12:00:00Z`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			var lastSync *timestamppb.Timestamp

			if lastSyncFile != "" {
				data, err := os.ReadFile(lastSyncFile)
				if err != nil {
					return err
				}

				timeStr := strings.TrimSpace(string(data))
				t, err := time.Parse(time.RFC3339, timeStr)
				if err != nil {
					return err
				}
				lastSync = timestamppb.New(t)
			} else if lastSyncTime != "" {
				t, err := time.Parse(time.RFC3339, lastSyncTime)
				if err != nil {
					return err
				}
				lastSync = timestamppb.New(t)
			}

			if err := app.Sync(ctx, lastSync); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&lastSyncFile, "file", "f", "", "Файл с временем последней синхронизации (формат RFC3339)")
	cmd.Flags().StringVarP(&lastSyncTime, "time", "t", "", "Время последней синхронизации (формат RFC3339)")

	return cmd
}
