package server

import (
	"go-svc-gophkeeper/internal/auth"
	"go-svc-gophkeeper/internal/storage"
)

// Service объединяет все сервисы приложения
type Service struct {
	Auth   *AuthService
	Secret *SecretService
}

// NewService создает новый экземпляр сервиса
func NewService(userRepo storage.UserRepository, secretRepo storage.SecretRepository, encryptor Encryptor, jwtManager *auth.JWTManager) *Service {
	return &Service{
		Auth:   NewAuthService(userRepo, jwtManager),
		Secret: NewSecretService(secretRepo, encryptor),
	}
}

// UserRepository интерфейс шифрования
type Encryptor interface {
	Encrypt(data []byte) ([]byte, error)
	Decrypt(encryptedData []byte) ([]byte, error)
}
