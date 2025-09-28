package server

import (
	"context"
	"fmt"
	"time"

	v1 "go-svc-gophkeeper/gen/go/v1"
	"go-svc-gophkeeper/internal/app/server"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// AuthHandler обрабатывает gRPC запросы для аутентификации
type AuthHandler struct {
	v1.UnimplementedAuthServiceServer
	authService *server.AuthService
	tokenExpiry time.Duration
}

// NewAuthHandler создает новый экземпляр AuthHandler
func NewAuthHandler(authService *server.AuthService, tokenExpiry time.Duration) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		tokenExpiry: tokenExpiry,
	}
}

// Register обрабатывает запрос на регистрацию пользователя
func (h *AuthHandler) Register(ctx context.Context, req *v1.RegisterRequest) (*v1.RegisterResponse, error) {
	if err := validateRegisterRequest(req); err != nil {
		return nil, err
	}

	user, token, err := h.authService.Register(ctx, req.GetLogin(), req.GetPassword())
	if err != nil {
		return nil, mapAuthErrorToGRPC(err)
	}

	return &v1.RegisterResponse{
		UserId:      fmt.Sprintf("%d", user.ID),
		AccessToken: token,
	}, nil
}

// Login обрабатывает запрос на аутентификацию пользователя
func (h *AuthHandler) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginResponse, error) {
	if err := validateLoginRequest(req); err != nil {
		return nil, err
	}

	user, token, err := h.authService.Login(ctx, req.GetLogin(), req.GetPassword())
	if err != nil {
		return nil, mapAuthErrorToGRPC(err)
	}

	expiresAt := time.Now().Add(h.tokenExpiry)

	return &v1.LoginResponse{
		AccessToken: token,
		UserId:      fmt.Sprintf("%d", user.ID),
		ExpiresAt:   timestamppb.New(expiresAt),
	}, nil
}

// validateRegisterRequest проверяет корректность запроса регистрации
func validateRegisterRequest(req *v1.RegisterRequest) error {
	if req.GetLogin() == "" {
		return status.Error(codes.InvalidArgument, "login is required")
	}
	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}
	if len(req.GetPassword()) < 8 {
		return status.Error(codes.InvalidArgument, "password must be at least 8 characters long")
	}
	return nil
}

// validateLoginRequest проверяет корректность запроса логина
func validateLoginRequest(req *v1.LoginRequest) error {
	if req.GetLogin() == "" {
		return status.Error(codes.InvalidArgument, "login is required")
	}
	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}
	return nil
}

// mapAuthErrorToGRPC преобразует ошибки бизнес-логики в gRPC статусы
func mapAuthErrorToGRPC(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case err.Error() == "user already exists":
		return status.Error(codes.AlreadyExists, "user with this login already exists")
	case err.Error() == "invalid credentials":
		return status.Error(codes.Unauthenticated, "invalid login or password")
	case err.Error() == "login must be between 3 and 50 characters" ||
		err.Error() == "password must be at least 8 characters":
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
