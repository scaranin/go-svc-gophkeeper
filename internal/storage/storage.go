package storage

import (
	"context"

	"go-svc-gophkeeper/internal/models"
)

// Storage объединяет все интерфейсы репозиториев.
type Storage interface {
	User() UserRepository
	Secret() SecretRepository
	Close()
}

// UserRepository определяет контракт для работы с данными пользователей.
type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) (int, error)
	GetUserByID(ctx context.Context, id int) (*models.User, error)
	GetUserByLogin(ctx context.Context, login string) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, id int) error
}

// SecretRepository определяет контракт для работы с секретами.
type SecretRepository interface {
	CreateSecret(ctx context.Context, secret *models.Secret) (int, error)
	GetSecretByID(ctx context.Context, id, userID int) (*models.Secret, error)
	ListByUserID(ctx context.Context, userID int) ([]*models.Secret, error)
	UpdateSecret(ctx context.Context, secret *models.Secret) error
	DeleteSecret(ctx context.Context, id, userID int) error
}
