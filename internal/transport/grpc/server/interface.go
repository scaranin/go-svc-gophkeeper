package server

import (
	"context"

	v1 "go-svc-gophkeeper/gen/go/v1"
	"go-svc-gophkeeper/internal/auth"

	"google.golang.org/grpc"
)

// Service определяет интерфейс для бизнес-сервисов
type Service interface {
	Run() error
	Close() error
}

// GRPCHandler определяет интерфейс для gRPC хендлеров
type GRPCHandler interface {
	Register(server *grpc.Server)
}

// AuthService интерфейс для сервиса аутентификации
type AuthService interface {
	Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginResponse, error)
	Register(ctx context.Context, req *v1.RegisterRequest) (*v1.RegisterResponse, error)
	ValidateToken(ctx context.Context, token string) (*auth.Claims, error)
}

// SecretService интерфейс для сервиса секретов
type SecretService interface {
	StoreSecret(ctx context.Context, req *v1.CreateSecretRequest) (*v1.CreateSecretResponse, error)
	GetSecret(ctx context.Context, req *v1.GetSecretRequest) (*v1.GetSecretResponse, error)
	ListSecrets(ctx context.Context, req *v1.ListSecretsRequest) (*v1.ListSecretsResponse, error)
	DeleteSecret(ctx context.Context, req *v1.DeleteSecretRequest) (*v1.DeleteSecretResponse, error)
}
