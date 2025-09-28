package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go-svc-gophkeeper/internal/app/server"
	"go-svc-gophkeeper/internal/auth"
	"go-svc-gophkeeper/internal/config"
	"go-svc-gophkeeper/internal/storage/postgres"
	grpcserver "go-svc-gophkeeper/internal/transport/grpc/server"
)

// Dependencies содержит все зависимости приложения
type Dependencies struct {
	Store       *postgres.Store
	Services    *server.Service
	AuthHandler *grpcserver.AuthHandler
	// В будущем добавим:
	// SecretHandler *grpcserver.SecretHandler
	// JWTManager    *auth.JWTManager
}

// Cleanup освобождает ресурсы зависимостей
func (d *Dependencies) Cleanup() {
	if d.Store != nil {
		d.Store.Close()
		log.Println("Storage connection closed")
	}
}

// setupDependencies создает и инициализирует все зависимости приложения
func setupDependencies(cfg *config.Config) (*Dependencies, error) {
	// 1. Инициализация хранилища
	store, err := postgres.New(cfg.Database.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	// 2. Проверка подключения к БД
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := store.Ping(ctx); err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	// 3. Запуск миграций БД
	if err := store.Migrate(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// 4. Инициализация JWT менеджера
	jwtManager, err := auth.NewJWTManager(cfg.Auth.JWTSecret, cfg.Auth.AccessTokenExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize JWT manager: %w", err)
	}

	// 5. Инициализация сервиса шифрования
	encryptor, err := server.NewEncryptionService(cfg.Encryption.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize encryption service: %w", err)
	}

	// 6. Создание бизнес-сервисов
	services := server.NewService(
		store.User(),   // UserRepository
		store.Secret(), // SecretRepository
		encryptor,      // Encryptor
		jwtManager,     // JWTManager
	)

	// 7. Создание gRPC хендлеров
	authHandler := grpcserver.NewAuthHandler(services.Auth, cfg.Auth.AccessTokenExpiry)

	return &Dependencies{
		Store:       store,
		Services:    services,
		AuthHandler: authHandler,
	}, nil
}
