package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-svc-gophkeeper/internal/app/server"
	"go-svc-gophkeeper/internal/auth"
	"go-svc-gophkeeper/internal/config"
	"go-svc-gophkeeper/internal/storage/postgres"

	v1 "go-svc-gophkeeper/gen/go/v1"

	"google.golang.org/grpc"
)

// dependencies содержит все зависимости приложения
type dependencies struct {
	store             *postgres.Store
	jwtManager        *auth.JWTManager
	encryptionService *server.EncryptionService
	appServices       *server.Service
}

// RunServer запускает gRPC сервер
func RunServer() {
	cfg, err := config.Load("./configs/server.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	deps, err := setupDependencies(cfg)
	if err != nil {
		log.Fatalf("Failed to setup dependencies: %v", err)
	}
	defer deps.cleanup()

	log.Println("All dependencies initialized successfully")

	grpcServer, err := setupGRPCServer(deps, cfg)
	if err != nil {
		log.Fatalf("Failed to setup gRPC server: %v", err)
	}

	if err := runServer(grpcServer, cfg); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}

// cleanup освобождает ресурсы
func (d *dependencies) cleanup() {
	if d.store != nil {
		d.store.Close()
	}
}

// setupDependencies создает все зависимости приложения
func setupDependencies(cfg *config.Config) (*dependencies, error) {
	store, err := createStorage(cfg)
	if err != nil {
		return nil, err
	}

	jwtManager, err := createJWTManager(cfg)
	if err != nil {
		store.Close()
		return nil, err
	}

	encryptionService, err := createEncryptionService(cfg)
	if err != nil {
		store.Close()
		return nil, err
	}

	appServices, err := createAppServices(store, encryptionService, jwtManager)
	if err != nil {
		store.Close()
		return nil, err
	}

	return &dependencies{
		store:             store,
		jwtManager:        jwtManager,
		encryptionService: encryptionService,
		appServices:       appServices,
	}, nil
}

// createStorage создает хранилище
func createStorage(cfg *config.Config) (*postgres.Store, error) {
	store, err := postgres.New(cfg.Database.DSN, cfg.Database.MigrationsPath)
	if err != nil {
		return nil, fmt.Errorf("create storage: %w", err)
	}
	log.Printf("Connected to database, migrations path: %s", cfg.Database.MigrationsPath)
	return store, nil
}

// createJWTManager создает JWT менеджер
func createJWTManager(cfg *config.Config) (*auth.JWTManager, error) {
	jwtManager, err := auth.NewJWTManager(cfg.Auth.JWTSecret, cfg.Auth.AccessTokenExpiry)
	if err != nil {
		return nil, err
	}
	log.Println("JWT manager initialized")
	return jwtManager, nil
}

// createEncryptionService создает сервис шифрования
func createEncryptionService(cfg *config.Config) (*server.EncryptionService, error) {
	encryptionService, err := server.NewEncryptionService(cfg.Encryption.Key)
	if err != nil {
		return nil, err
	}
	log.Println("Encryption service initialized")
	return encryptionService, nil
}

// createAppServices создает бизнес-сервисы приложения
func createAppServices(store *postgres.Store, encryptionService *server.EncryptionService, jwtManager *auth.JWTManager) (*server.Service, error) {
	appServices := server.NewService(store.User(), store.Secret(), encryptionService, jwtManager)
	log.Println("Application services initialized")
	return appServices, nil
}

// setupGRPCServer создает gRPC сервер
func setupGRPCServer(deps *dependencies, cfg *config.Config) (*grpc.Server, error) {
	authHandler := NewAuthHandler(deps.appServices.Auth, cfg.Auth.AccessTokenExpiry)
	secretHandler := NewSecretHandler(deps.appServices.Secret)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(JWTAuthMiddleware(deps.appServices.Auth)),
	)

	v1.RegisterAuthServiceServer(grpcServer, authHandler)
	v1.RegisterSecretServiceServer(grpcServer, secretHandler)

	log.Println("gRPC server configured")
	return grpcServer, nil
}

// runServer запускает сервер
func runServer(grpcServer *grpc.Server, cfg *config.Config) error {
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	log.Printf("Starting gRPC server on %s", addr)

	serverErr := make(chan error, 1)
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			serverErr <- err
		}
	}()

	return waitForShutdown(grpcServer, cfg.Server.ShutdownTimeout, serverErr)
}

// waitForShutdown ожидает сигналов завершения
func waitForShutdown(grpcServer *grpc.Server, timeout time.Duration, serverErr chan error) error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		log.Printf("Received signal: %v. Starting graceful shutdown...", sig)
		return shutdownServer(grpcServer, timeout)

	case err := <-serverErr:
		log.Printf("Server error: %v", err)
		return err
	}
}

// shutdownServer выполняет graceful shutdown
func shutdownServer(grpcServer *grpc.Server, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	stopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		log.Println("gRPC server stopped gracefully")
		return nil

	case <-ctx.Done():
		log.Printf("Shutdown timeout exceeded. Forcing stop.")
		grpcServer.Stop()
		return fmt.Errorf("shutdown timeout exceeded")
	}
}
