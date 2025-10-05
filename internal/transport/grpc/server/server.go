package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"go-svc-gophkeeper/internal/app/server"
	"go-svc-gophkeeper/internal/auth"
	"go-svc-gophkeeper/internal/config"
	"go-svc-gophkeeper/internal/storage/postgres"

	v1 "go-svc-gophkeeper/gen/go/v1"

	"google.golang.org/grpc"
)

// Server представляет основной сервер приложения
type Server struct {
	config     *config.Config
	grpcServer *grpc.Server
	listener   net.Listener
	store      *postgres.Store
	services   *server.Service
}

// New создает новый экземпляр сервера
func New(cfg *config.Config) (*Server, error) {
	s := &Server{
		config: cfg,
	}

	if err := s.initDependencies(); err != nil {
		return nil, err
	}

	if err := s.setupGRPCServer(); err != nil {
		s.cleanup()
		return nil, err
	}

	return s, nil
}

// initDependencies инициализирует все зависимости сервера
func (s *Server) initDependencies() error {
	store, err := postgres.New(s.config.Database.DSN, s.config.Database.MigrationsPath)
	if err != nil {
		return err
	}
	s.store = store

	jwtManager, err := auth.NewJWTManager(s.config.Auth.JWTSecret, s.config.Auth.AccessTokenExpiry)
	if err != nil {
		return err
	}

	encryptionService, err := server.NewEncryptionService(s.config.Encryption.Key)
	if err != nil {
		return err
	}

	appServices := server.NewService(store.User(), store.Secret(), encryptionService, jwtManager)
	s.services = appServices

	return nil
}

// setupGRPCServer настраивает gRPC сервер
func (s *Server) setupGRPCServer() error {
	authHandler := s.newAuthHandler()
	secretHandler := s.newSecretHandler()

	s.grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(s.jwtAuthMiddleware()),
	)

	v1.RegisterAuthServiceServer(s.grpcServer, authHandler)
	v1.RegisterSecretServiceServer(s.grpcServer, secretHandler)

	return nil
}

// newAuthHandler создает хендлер для аутентификации
func (s *Server) newAuthHandler() *AuthHandler {
	return &AuthHandler{
		authService: s.services.Auth,
		tokenExpiry: s.config.Auth.AccessTokenExpiry,
	}
}

// newSecretHandler создает хендлер для работы с секретами
func (s *Server) newSecretHandler() *SecretHandler {
	return &SecretHandler{
		secretService: s.services.Secret,
	}
}

// jwtAuthMiddleware создает middleware для JWT аутентификации
func (s *Server) jwtAuthMiddleware() grpc.UnaryServerInterceptor {
	return JWTAuthMiddleware(s.services.Auth)
}

// Run запускает сервер
func (s *Server) Run() error {
	addr := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.listener = listener

	serverErr := make(chan error, 1)
	go func() {
		if err := s.grpcServer.Serve(s.listener); err != nil {
			serverErr <- err
		}
	}()

	return s.waitForShutdown(serverErr)
}

// waitForShutdown ожидает сигналов завершения работы
func (s *Server) waitForShutdown(serverErr chan error) error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		log.Printf("Received signal: %v. Starting graceful shutdown...", sig)
		return s.shutdown()

	case err := <-serverErr:
		log.Println(err)
		return err
	}
}

// shutdown выполняет graceful shutdown сервера
func (s *Server) shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.config.Server.ShutdownTimeout)
	defer cancel()

	stopped := make(chan struct{})
	go func() {
		s.grpcServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		log.Println("gRPC server stopped gracefully")
		return nil

	case <-ctx.Done():
		log.Printf("Shutdown timeout exceeded. Forcing stop.")
		s.grpcServer.Stop()
		return fmt.Errorf("shutdown timeout exceeded")
	}
}

// Close освобождает ресурсы сервера
func (s *Server) Close() error {
	s.cleanup()
	return nil
}

// cleanup освобождает все ресурсы
func (s *Server) cleanup() {
	if s.store != nil {
		s.store.Close()
	}
}

// RunServer запускает сервер приложения
func RunServer() {
	cfg, err := config.Load("./configs/server.yaml")
	if err != nil {
		log.Fatal(err)
	}

	srv, err := New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer srv.Close()

	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}
}
