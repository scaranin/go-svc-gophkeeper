package server

import (
	"context"
	"fmt"

	v1 "go-svc-gophkeeper/gen/go/v1"
	"go-svc-gophkeeper/internal/models"
)

type SecretService struct {
	secretRepo SecretRepository
	encryptor  Encryptor
}

// NewSecretService создает новый сервис секретов
func NewSecretService(secretRepo SecretRepository, encryptor Encryptor) *SecretService {
	return &SecretService{
		secretRepo: secretRepo,
		encryptor:  encryptor,
	}
}

// CreateSecret создает новый секрет
func (s *SecretService) CreateSecret(ctx context.Context, userID int, secretType models.SecretType, name string, data []byte, metadata *v1.Metadata) (*models.Secret, error) {
	encryptedData, err := s.encryptor.Encrypt(data)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt data: %w", err)
	}

	metadataWrapper := models.FromProto(metadata)
	metadataBytes, err := metadataWrapper.ToBytes()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize metadata: %w", err)
	}

	secret := &models.Secret{
		UserID:        userID,
		Type:          secretType,
		Name:          name,
		EncryptedData: encryptedData,
		Metadata:      metadataBytes,
		Version:       1,
	}

	secretID, err := s.secretRepo.CreateSecret(ctx, secret)
	if err != nil {
		return nil, fmt.Errorf("failed to create secret: %w", err)
	}

	secret.ID = secretID
	return secret, nil
}

// GetSecret возвращает секрет по ID с metadata
func (s *SecretService) GetSecret(ctx context.Context, secretID, userID int) (*models.Secret, *v1.Metadata, error) {
	secret, err := s.secretRepo.GetSecretByID(ctx, secretID, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get secret: %w", err)
	}

	if secret == nil {
		return nil, nil, nil
	}

	decryptedData, err := s.encryptor.Decrypt(secret.EncryptedData)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	var metadataWrapper models.MetadataWrapper
	if err := metadataWrapper.FromBytes(secret.Metadata); err != nil {
		return nil, nil, fmt.Errorf("failed to deserialize metadata: %w", err)
	}

	result := *secret
	result.EncryptedData = decryptedData

	return &result, metadataWrapper.ToProto(), nil
}

// UpdateSecret обновляет существующий секрет
func (s *SecretService) UpdateSecret(ctx context.Context, secretID, userID int, name string, data []byte, metadata *v1.Metadata, version int) (*models.Secret, error) {
	currentSecret, err := s.secretRepo.GetSecretByID(ctx, secretID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}
	if currentSecret == nil {
		return nil, fmt.Errorf("secret not found")
	}

	if currentSecret.Version != version {
		return nil, fmt.Errorf("version conflict")
	}

	encryptedData, err := s.encryptor.Encrypt(data)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt data: %w", err)
	}

	metadataWrapper := models.FromProto(metadata)
	metadataBytes, err := metadataWrapper.ToBytes()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize metadata: %w", err)
	}

	updatedSecret := &models.Secret{
		ID:            secretID,
		UserID:        userID,
		Type:          currentSecret.Type,
		Name:          name,
		EncryptedData: encryptedData,
		Metadata:      metadataBytes,
		Version:       version + 1,
	}

	err = s.secretRepo.UpdateSecret(ctx, updatedSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to update secret: %w", err)
	}

	return updatedSecret, nil
}

// DeleteSecret помечает секрет как удаленный
func (s *SecretService) DeleteSecret(ctx context.Context, secretID, userID int) error {
	err := s.secretRepo.DeleteSecret(ctx, secretID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}
	return nil
}

// ListSecrets возвращает все секреты пользователя
func (s *SecretService) ListSecrets(ctx context.Context, userID int) ([]*models.Secret, error) {
	return s.secretRepo.ListByUserID(ctx, userID)
}

// SyncSecrets синхронизирует секреты клиента с сервером
func (s *SecretService) SyncSecrets(ctx context.Context, userID int, clientSecrets []*models.Secret) ([]*models.Secret, error) {
	serverSecrets, err := s.secretRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return serverSecrets, nil
}
