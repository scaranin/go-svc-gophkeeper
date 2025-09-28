package server

import (
	"context"
	"fmt"

	"go-svc-gophkeeper/internal/models"
)

type SecretService struct {
	secretRepo SecretRepository
	encryptor  Encryptor
}

func NewSecretService(secretRepo SecretRepository, encryptor Encryptor) *SecretService {
	return &SecretService{
		secretRepo: secretRepo,
		encryptor:  encryptor,
	}
}

// CreateSecret создает новый секрет
func (s *SecretService) CreateSecret(ctx context.Context, userID int, secretType models.SecretType, name string, data []byte, metadata []byte) (*models.Secret, error) {
	encryptedData, err := s.encryptor.Encrypt(data)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt data: %w", err)
	}

	secret := &models.Secret{
		UserID:        userID,
		Type:          secretType,
		Name:          name,
		EncryptedData: encryptedData,
		Metadata:      metadata,
		Version:       1,
	}

	secretID, err := s.secretRepo.CreateSecret(ctx, secret)
	if err != nil {
		return nil, fmt.Errorf("failed to create secret: %w", err)
	}

	secret.ID = secretID
	return secret, nil
}

// GetSecret возвращает секрет по ID
func (s *SecretService) GetSecret(ctx context.Context, secretID, userID int) (*models.Secret, error) {
	secret, err := s.secretRepo.GetSecretByID(ctx, secretID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	decryptedData, err := s.encryptor.Decrypt(secret.EncryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	result := *secret
	result.EncryptedData = decryptedData

	return &result, nil
}

// ListSecrets возвращает все секреты пользователя
func (s *SecretService) ListSecrets(ctx context.Context, userID int) ([]*models.Secret, error) {
	return s.secretRepo.ListByUserID(ctx, userID)
}

// SyncSecrets синхронизирует секреты клиента с сервером
func (s *SecretService) SyncSecrets(ctx context.Context, userID int, clientSecrets []*models.Secret) ([]*models.Secret, error) {
	// TODO: Реализовать логику синхронизации
	// - Сравнение версий
	// - Разрешение конфликтов
	// - Возврат измененных секретов

	serverSecrets, err := s.secretRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return serverSecrets, nil
}
