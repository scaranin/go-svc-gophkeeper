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
	"google.golang.org/grpc/reflection"
)

// RunServer запускает gRPC сервер
func RunServer() {
	cfg, err := config.Load("./configs/server.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	store, err := postgres.New(cfg.Database.DSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer store.Close()

	log.Println("Connected to database successfully")

	jwtManager, err := auth.NewJWTManager(cfg.Auth.JWTSecret, cfg.Auth.AccessTokenExpiry)
	if err != nil {
		log.Fatalf("Failed to create JWT manager: %v", err)
	}

	encryptionService, err := server.NewEncryptionService(cfg.Encryption.Key)
	if err != nil {
		log.Fatalf("Failed to create encryption service: %v", err)
	}

	appServices := server.NewService(store.User(), store.Secret(), encryptionService, jwtManager)

	authHandler := NewAuthHandler(appServices.Auth, cfg.Auth.AccessTokenExpiry)
	secretHandler := NewSecretHandler(appServices.Secret)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(JWTAuthMiddleware(appServices.Auth)),
	)

	v1.RegisterAuthServiceServer(grpcServer, authHandler)
	v1.RegisterSecretServiceServer(grpcServer, secretHandler)

	reflection.Register(grpcServer)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", addr, err)
	}

	log.Printf("Starting gRPC server on %s", addr)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve gRPC server: %v", err)
		}
	}()

	waitForShutdown(grpcServer, cfg.Server.ShutdownTimeout)
}

// waitForShutdown ожидает сигналов завершения
func waitForShutdown(grpcServer *grpc.Server, timeout time.Duration) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	log.Printf("Received signal: %v. Starting graceful shutdown...", sig)

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
	case <-ctx.Done():
		log.Printf("Shutdown timeout exceeded. Forcing stop.")
		grpcServer.Stop()
	}

	log.Println("Server shutdown completed")
}
