package client

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	v1 "go-svc-gophkeeper/gen/go/v1"
	"go-svc-gophkeeper/internal/config"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// App представляет основное клиентское приложение
type App struct {
	config       *config.ClientConfig
	conn         *grpc.ClientConn
	authClient   v1.AuthServiceClient
	secretClient v1.SecretServiceClient
	token        string
}

// NewApp создает новый экземпляр клиентского приложения
func NewApp(cfg *config.ClientConfig) (*App, error) {
	app := &App{
		config: cfg,
	}

	if err := app.ensureDataDir(); err != nil {
		return nil, err
	}

	if err := app.connect(); err != nil {
		return nil, err
	}

	log.Printf("Клиент инициализирован: сервер=%s, таймаут=%v", cfg.ServerURL, cfg.Timeout)
	return app, nil
}

// ensureDataDir создает директорию для данных, если она не существует
func (a *App) ensureDataDir() error {
	dir := a.config.DataDir

	if dir == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		dir = home
	} else if strings.HasPrefix(dir, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		relativePath := strings.TrimPrefix(dir, "~/")
		dir = filepath.Join(home, relativePath)
	}

	dir = filepath.Clean(dir)
	a.config.DataDir = dir

	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	return nil
}

// connect устанавливает соединение с gRPC сервером
func (a *App) connect() error {
	conn, err := grpc.NewClient(a.config.ServerURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}

	a.conn = conn
	a.authClient = v1.NewAuthServiceClient(conn)
	a.secretClient = v1.NewSecretServiceClient(conn)

	log.Printf("Успешно подключено к серверу: %s", a.config.ServerURL)
	return nil
}

// createContext создает контекст с таймаутом
func (a *App) createContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, a.config.Timeout)
}

// createAuthContext создает контекст с авторизационным токеном
func (a *App) createAuthContext(ctx context.Context) context.Context {
	if a.token == "" {
		return ctx
	}
	return metadata.NewOutgoingContext(ctx, metadata.Pairs(
		"authorization", "Bearer "+a.token,
	))
}

// Register регистрирует нового пользователя
func (a *App) Register(ctx context.Context, login, password string) error {
	fmt.Printf("Регистрация пользователя: %s\n", login)

	ctx, cancel := a.createContext(ctx)
	defer cancel()

	resp, err := a.authClient.Register(ctx, &v1.RegisterRequest{
		Login:    login,
		Password: password,
	})
	if err != nil {
		return err
	}

	a.token = resp.AccessToken
	fmt.Printf("Регистрация успешна: UserID=%s\n", resp.UserId)
	return nil
}

// Login выполняет аутентификацию пользователя
func (a *App) Login(ctx context.Context, login, password string) error {
	fmt.Printf("Вход пользователя: %s\n", login)

	ctx, cancel := a.createContext(ctx)
	defer cancel()

	resp, err := a.authClient.Login(ctx, &v1.LoginRequest{
		Login:    login,
		Password: password,
	})
	if err != nil {
		return err
	}

	a.token = resp.AccessToken
	fmt.Printf("Вход выполнен успешно: UserID=%s\n", resp.UserId)
	return nil
}

// CreateSecret создает новый секрет
func (a *App) CreateSecret(ctx context.Context, secretType v1.SecretType, name string, metadata *v1.Metadata, encryptedData []byte) error {
	fmt.Printf("Создание секрета: %s\n", name)

	ctx, cancel := a.createContext(ctx)
	defer cancel()

	authCtx := a.createAuthContext(ctx)
	resp, err := a.secretClient.CreateSecret(authCtx, &v1.CreateSecretRequest{
		Type:          secretType,
		Name:          name,
		Metadata:      metadata,
		EncryptedData: encryptedData,
	})
	if err != nil {
		return err
	}

	fmt.Printf("Секрет создан успешно: SecretID=%s, Версия=%d\n", resp.SecretId, resp.Version)
	return nil
}

// ListSecrets выводит список всех секретов
func (a *App) ListSecrets(ctx context.Context, includeDeleted bool) error {
	fmt.Println("Получение списка секретов")

	ctx, cancel := a.createContext(ctx)
	defer cancel()

	authCtx := a.createAuthContext(ctx)
	resp, err := a.secretClient.ListSecrets(authCtx, &v1.ListSecretsRequest{
		IncludeDeleted: includeDeleted,
	})
	if err != nil {
		return err
	}

	fmt.Printf("Найдено секретов: %d\n", len(resp.Secrets))
	for i, secret := range resp.Secrets {
		fmt.Printf("%d. ID=%s, Тип=%s, Имя=%s, Версия=%d\n",
			i+1, secret.Id, secret.Type.String(), secret.Name, secret.Version)
	}
	return nil
}

// GetSecret выводит информацию о конкретном секрете
func (a *App) GetSecret(ctx context.Context, secretID string) error {
	fmt.Printf("Получение секрета: %s\n", secretID)

	ctx, cancel := a.createContext(ctx)
	defer cancel()

	authCtx := a.createAuthContext(ctx)
	resp, err := a.secretClient.GetSecret(authCtx, &v1.GetSecretRequest{
		SecretId: secretID,
	})
	if err != nil {
		return err
	}

	info := resp.Info
	fmt.Printf("Детали секрета:\n")
	fmt.Printf("  ID: %s\n", info.Id)
	fmt.Printf("  Тип: %s\n", info.Type.String())
	fmt.Printf("  Имя: %s\n", info.Name)
	fmt.Printf("  Версия: %d\n", info.Version)
	if info.Metadata != nil {
		fmt.Printf("  Название: %s\n", info.Metadata.Name)
		fmt.Printf("  Описание: %s\n", info.Metadata.Description)
		fmt.Printf("  Веб-сайт: %s\n", info.Metadata.Website)
		fmt.Printf("  Теги: %s\n", info.Metadata.Tags)
	}
	fmt.Printf("  Создан: %v\n", info.CreatedAt.AsTime().Format(time.RFC3339))
	fmt.Printf("  Обновлен: %v\n", info.UpdatedAt.AsTime().Format(time.RFC3339))

	if len(resp.EncryptedData) > 0 {
		fmt.Printf("  Зашифрованные данные: %d байт\n", len(resp.EncryptedData))
	}

	return nil
}

// UpdateSecret обновляет существующий секрет
func (a *App) UpdateSecret(ctx context.Context, secretID string, encryptedData []byte) error {
	fmt.Printf("Обновление секрета: %s\n", secretID)

	ctx, cancel := a.createContext(ctx)
	defer cancel()

	authCtx := a.createAuthContext(ctx)
	resp, err := a.secretClient.UpdateSecret(authCtx, &v1.UpdateSecretRequest{
		SecretId:      secretID,
		EncryptedData: encryptedData,
	})
	if err != nil {
		return err
	}

	fmt.Printf("Секрет обновлен успешно: SecretID=%s, Версия=%d\n", secretID, resp.NewVersion)
	return nil
}

// DeleteSecret удаляет секрет
func (a *App) DeleteSecret(ctx context.Context, secretID string) error {
	fmt.Printf("Удаление секрета: %s\n", secretID)

	ctx, cancel := a.createContext(ctx)
	defer cancel()

	authCtx := a.createAuthContext(ctx)
	_, err := a.secretClient.DeleteSecret(authCtx, &v1.DeleteSecretRequest{
		SecretId: secretID,
	})
	if err != nil {
		return err
	}

	fmt.Printf("Секрет удален успешно: %s\n", secretID)
	return nil
}

// Sync синхронизирует данные с сервером
func (a *App) Sync(ctx context.Context, lastSync *timestamppb.Timestamp) error {
	fmt.Println("Синхронизация с сервером")

	ctx, cancel := a.createContext(ctx)
	defer cancel()

	authCtx := a.createAuthContext(ctx)
	resp, err := a.secretClient.Sync(authCtx, &v1.SyncRequest{
		LastSync: lastSync,
	})
	if err != nil {
		return err
	}

	fmt.Printf("Синхронизация выполнена успешно: %d секретов\n", resp.SyncState.TotalSecrets)
	if resp.SyncState.LastSync != nil {
		fmt.Printf("Последняя синхронизация: %v\n", resp.SyncState.LastSync.AsTime().Format(time.RFC3339))
	}

	return nil
}

// GetToken возвращает текущий токен аутентификации
func (a *App) GetToken() string {
	return a.token
}

// SetToken устанавливает токен аутентификации
func (a *App) SetToken(token string) {
	a.token = token
}

// IsAuthenticated проверяет, аутентифицирован ли пользователь
func (a *App) IsAuthenticated() bool {
	return a.token != ""
}

// Close освобождает ресурсы приложения
func (a *App) Close() {
	if a.conn != nil {
		a.conn.Close()
		log.Println("Соединение с сервером закрыто")
	}
}
