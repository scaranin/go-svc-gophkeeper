package server

import (
	"context"

	"go-svc-gophkeeper/internal/auth"
	"go-svc-gophkeeper/internal/models"
)

// Service объединяет все сервисы приложения
type Service struct {
	Auth   *AuthService
	Secret *SecretService
}

// NewService создает новый экземпляр сервиса
func NewService(userRepo UserRepository, secretRepo SecretRepository, encryptor Encryptor, jwtManager *auth.JWTManager) *Service {
	return &Service{
		Auth:   NewAuthService(userRepo, jwtManager),
		Secret: NewSecretService(secretRepo, encryptor),
	}
}

// Интерфейсы для зависимостей
type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) (int, error)
	GetUserByLogin(ctx context.Context, login string) (*models.User, error)
	GetUserByID(ctx context.Context, id int) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, id int) error
}

type SecretRepository interface {
	CreateSecret(ctx context.Context, secret *models.Secret) (int, error)
	GetSecretByID(ctx context.Context, id, userID int) (*models.Secret, error)
	ListByUserID(ctx context.Context, userID int) ([]*models.Secret, error)
	UpdateSecret(ctx context.Context, secret *models.Secret) error
	DeleteSecret(ctx context.Context, id, userID int) error
}

type Encryptor interface {
	Encrypt(data []byte) ([]byte, error)
	Decrypt(encryptedData []byte) ([]byte, error)
}
