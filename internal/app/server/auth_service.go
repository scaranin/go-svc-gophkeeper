package server

import (
	"context"
	"fmt"

	"go-svc-gophkeeper/internal/models"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo UserRepository
}

func NewAuthService(userRepo UserRepository) *AuthService {
	return &AuthService{
		userRepo: userRepo,
	}
}

// Register регистрирует нового пользователя
func (s *AuthService) Register(ctx context.Context, login, password string) (*models.User, error) {
	if len(login) < 5 || len(login) > 15 {
		return nil, fmt.Errorf("login must be between 5 and 15 characters")
	}
	if len(password) < 8 {
		return nil, fmt.Errorf("password must be at least 8 characters")
	}

	existUser, err := s.userRepo.GetUserByLogin(ctx, login)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if existUser != nil {
		return nil, fmt.Errorf("user with login %s already exists", login)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		Login:        login,
		PasswordHash: string(passwordHash),
	}

	userID, err := s.userRepo.CreateUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	user.ID = userID
	return user, nil
}

// Login аутентифицирует пользователя
func (s *AuthService) Login(ctx context.Context, login, password string) (*models.User, error) {
	user, err := s.userRepo.GetUserByLogin(ctx, login)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	return user, nil
}

// GetUser возвращает пользователя по ID
func (s *AuthService) GetUser(ctx context.Context, userID int) (*models.User, error) {
	return s.userRepo.GetUserByID(ctx, userID)
}
