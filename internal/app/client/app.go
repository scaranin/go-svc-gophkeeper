package client

import (
	"context"
	"fmt"

	"go-svc-gophkeeper/internal/config"
)

// App представляет основное клиентское приложение
type App struct {
	config *config.ClientConfig
}

// NewApp создает новый экземпляр клиентского приложения
func NewApp(cfg *config.ClientConfig) (*App, error) {
	return &App{
		config: cfg,
	}, nil
}

// Register регистрирует нового пользователя
func (a *App) Register(ctx context.Context, login, password string) error {
	fmt.Printf("Регистрация пользователя: %s\n", login)

	return nil
}

// Login выполняет аутентификацию пользователя
func (a *App) Login(ctx context.Context, login, password string) error {
	fmt.Printf("Вход пользователя: %s\n", login)

	return nil
}

// CreateSecret создает новый секрет
func (a *App) CreateSecret(ctx context.Context) error {
	fmt.Println("Создание секрета")

	return nil
}

// ListSecrets выводит список всех секретов
func (a *App) ListSecrets(ctx context.Context) error {
	fmt.Println("Список секретов")

	return nil
}

// GetSecret выводит информацию о конкретном секрете
func (a *App) GetSecret(ctx context.Context, secretID string) error {
	fmt.Printf("Получение секрета: %s\n", secretID)

	return nil
}

// Sync синхронизирует данные с сервером
func (a *App) Sync(ctx context.Context) error {
	fmt.Println("Синхронизация с сервером")

	return nil
}

// Close освобождает ресурсы приложения
func (a *App) Close() {

}
