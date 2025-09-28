package main

import (
	"go-svc-gophkeeper/internal/app/server"
	"go-svc-gophkeeper/internal/auth"
	"go-svc-gophkeeper/internal/config"
	"go-svc-gophkeeper/internal/storage/postgres"
	grpchandler "go-svc-gophkeeper/internal/transport/grpc/server"

	v1 "go-svc-gophkeeper/gen/go/v1"

	"google.golang.org/grpc"
)

// setupServices инициализирует все зависимости приложения
func setupServices(cfg *config.Config) (*grpc.Server, error) {
	store, err := postgres.New(cfg.Database.DSN)
	if err != nil {
		return nil, err
	}

	jwtManager, err := auth.NewJWTManager(cfg.Auth.JWTSecret, cfg.Auth.AccessTokenExpiry)
	if err != nil {
		store.Close()
		return nil, err
	}

	encryptionService, err := server.NewEncryptionService(cfg.Encryption.Key)
	if err != nil {
		store.Close()
		return nil, err
	}

	appService := server.NewService(store.User(), store.Secret(), encryptionService, jwtManager)

	authHandler := grpchandler.NewAuthHandler(appService.Auth, cfg.Auth.AccessTokenExpiry)
	secretHandler := grpchandler.NewSecretHandler(appService.Secret)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpchandler.JWTAuthMiddleware(appService.Auth)),
	)

	v1.RegisterAuthServiceServer(grpcServer, authHandler)
	v1.RegisterSecretServiceServer(grpcServer, secretHandler)

	return grpcServer, nil
}
