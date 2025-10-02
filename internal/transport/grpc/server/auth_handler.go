package server

import (
	"context"
	"fmt"
	"time"

	v1 "go-svc-gophkeeper/gen/go/v1"
	"go-svc-gophkeeper/internal/app/server"
	"go-svc-gophkeeper/internal/errors"

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
		return nil, mapAppErrorToGRPC(err)
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
		return nil, mapAppErrorToGRPC(err)
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

// mapAppErrorToGRPC преобразует ошибки приложения в gRPC статусы
func mapAppErrorToGRPC(err error) error {
	if err == nil {
		return nil
	}

	appErr, ok := err.(*errors.AppError)
	if !ok {
		return status.Error(codes.Internal, "internal server error")
	}

	switch appErr.Code {
	case "VALIDATION_ERROR", "INVALID_INPUT":
		return status.Error(codes.InvalidArgument, appErr.Message)
	case "ALREADY_EXISTS":
		return status.Error(codes.AlreadyExists, appErr.Message)
	case "INVALID_CREDENTIALS", "UNAUTHORIZED", "TOKEN_EXPIRED", "TOKEN_INVALID":
		return status.Error(codes.Unauthenticated, appErr.Message)
	case "FORBIDDEN":
		return status.Error(codes.PermissionDenied, appErr.Message)
	case "NOT_FOUND":
		return status.Error(codes.NotFound, appErr.Message)
	case "CONFLICT":
		return status.Error(codes.FailedPrecondition, appErr.Message)
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
