package server

import (
	"context"
	"fmt"
	"time"

	"go-svc-gophkeeper/internal/auth"
	"go-svc-gophkeeper/internal/errors"
	"go-svc-gophkeeper/internal/models"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo   UserRepository
	jwtManager *auth.JWTManager
}

// NewAuthService создание сервиса авторизации
func NewAuthService(userRepo UserRepository, jwtManager *auth.JWTManager) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

// Register регистрация пользователя
func (s *AuthService) Register(ctx context.Context, login, password string) (*models.User, string, error) {
	err := s.ValidateLogoPath(ctx, login, password)
	if err != nil {
		return nil, "", err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		Login:        login,
		PasswordHash: string(passwordHash),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	userID, err := s.userRepo.CreateUser(ctx, user)
	if err != nil {
		return nil, "", err
	}
	user.ID = userID

	token, err := s.jwtManager.GenerateToken(user.ID)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

// Login авторизация пользователя
func (s *AuthService) Login(ctx context.Context, login, password string) (*models.User, string, error) {
	if login == "" || password == "" {
		return nil, "", errors.ErrCredentialsRequired
	}

	user, err := s.userRepo.GetUserByLogin(ctx, login)
	if err != nil {
		return nil, "", err
	}
	if user == nil {
		return nil, "", errors.ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, "", errors.ErrInvalidCredentials
	}

	token, err := s.jwtManager.GenerateToken(user.ID)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

// ValidateToken проверка токена
func (s *AuthService) ValidateToken(token string) (int, error) {
	claims, err := s.jwtManager.ValidateToken(token)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}

// ValidateLogoPath проверка имени пользователя и пароля
func (s *AuthService) ValidateLogoPath(ctx context.Context, login, password string) error {
	if len(login) < 3 {
		return errors.ErrValidation
	}
	if len(password) < 8 {
		return errors.ErrValidation
	}

	existingUser, err := s.userRepo.GetUserByLogin(ctx, login)
	if err != nil {
		return err
	}

	if existingUser != nil {
		return errors.ErrAlreadyExists
	}
	return nil

}
