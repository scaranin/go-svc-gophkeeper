package server

import (
	"context"
	"testing"
	"time"

	v1 "go-svc-gophkeeper/gen/go/v1"
	"go-svc-gophkeeper/internal/models"

	"github.com/stretchr/testify/assert"
)

// MockSecretRepository мок для репозитория секретов
type MockSecretRepository struct {
	CreateSecretFunc  func(ctx context.Context, secret *models.Secret) (int, error)
	GetSecretByIDFunc func(ctx context.Context, id, userID int) (*models.Secret, error)
	ListByUserIDFunc  func(ctx context.Context, userID int) ([]*models.Secret, error)
	UpdateSecretFunc  func(ctx context.Context, secret *models.Secret) error
	DeleteSecretFunc  func(ctx context.Context, id, userID int) error
}

func (m *MockSecretRepository) CreateSecret(ctx context.Context, secret *models.Secret) (int, error) {
	if m.CreateSecretFunc != nil {
		return m.CreateSecretFunc(ctx, secret)
	}
	return 0, nil
}

func (m *MockSecretRepository) GetSecretByID(ctx context.Context, id, userID int) (*models.Secret, error) {
	if m.GetSecretByIDFunc != nil {
		return m.GetSecretByIDFunc(ctx, id, userID)
	}
	return nil, nil
}

func (m *MockSecretRepository) ListByUserID(ctx context.Context, userID int) ([]*models.Secret, error) {
	if m.ListByUserIDFunc != nil {
		return m.ListByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockSecretRepository) UpdateSecret(ctx context.Context, secret *models.Secret) error {
	if m.UpdateSecretFunc != nil {
		return m.UpdateSecretFunc(ctx, secret)
	}
	return nil
}

func (m *MockSecretRepository) DeleteSecret(ctx context.Context, id, userID int) error {
	if m.DeleteSecretFunc != nil {
		return m.DeleteSecretFunc(ctx, id, userID)
	}
	return nil
}

// MockEncryptor мок для шифратора
type MockEncryptor struct {
	EncryptFunc func(data []byte) ([]byte, error)
	DecryptFunc func(data []byte) ([]byte, error)
}

func (m *MockEncryptor) Encrypt(data []byte) ([]byte, error) {
	if m.EncryptFunc != nil {
		return m.EncryptFunc(data)
	}
	return data, nil
}

func (m *MockEncryptor) Decrypt(data []byte) ([]byte, error) {
	if m.DecryptFunc != nil {
		return m.DecryptFunc(data)
	}
	return data, nil
}

func TestSecretService_CreateSecret_Success(t *testing.T) {
	mockSecretRepo := &MockSecretRepository{
		CreateSecretFunc: func(ctx context.Context, secret *models.Secret) (int, error) {
			assert.Equal(t, 123, secret.UserID)
			assert.Equal(t, models.TypeLogin, secret.Type)
			assert.Equal(t, "Test Secret", secret.Name)
			assert.Equal(t, []byte("encrypted_data"), secret.EncryptedData)
			return 456, nil
		},
	}

	mockEncryptor := &MockEncryptor{
		EncryptFunc: func(data []byte) ([]byte, error) {
			return []byte("encrypted_data"), nil
		},
	}

	secretService := NewSecretService(mockSecretRepo, mockEncryptor)

	ctx := context.Background()
	metadata := &v1.Metadata{
		Name:        "Test Secret",
		Description: "Test Description",
		Website:     "https://test.com",
		Tags:        "test",
	}

	secret, err := secretService.CreateSecret(ctx, 123, models.TypeLogin, "Test Secret", []byte("test_data"), metadata)

	assert.NoError(t, err)
	assert.NotNil(t, secret)
	assert.Equal(t, 456, secret.ID)
	assert.Equal(t, 123, secret.UserID)
	assert.Equal(t, models.TypeLogin, secret.Type)
	assert.Equal(t, "Test Secret", secret.Name)
	assert.Equal(t, int32(1), secret.Version)
}

func TestSecretService_GetSecret_Success(t *testing.T) {
	mockSecretRepo := &MockSecretRepository{
		GetSecretByIDFunc: func(ctx context.Context, id, userID int) (*models.Secret, error) {
			assert.Equal(t, 456, id)
			assert.Equal(t, 123, userID)
			metadata := v1.Metadata{
				Name:        "Test Secret",
				Description: "Test Description",
				Website:     "https://test.com",
				Tags:        "test",
			}
			metadataWrapper := models.MetadataWrapper{Metadata: &metadata}
			metadataBytes, _ := metadataWrapper.ToBytes()

			return &models.Secret{
				ID:            456,
				UserID:        123,
				Type:          models.TypeLogin,
				Name:          "Test Secret",
				EncryptedData: []byte("encrypted_data"),
				Metadata:      metadataBytes,
				Version:       1,
				CreatedAt:     time.Now().Add(-24 * time.Hour),
				UpdatedAt:     time.Now(),
			}, nil
		},
	}

	mockEncryptor := &MockEncryptor{
		DecryptFunc: func(data []byte) ([]byte, error) {
			return []byte("decrypted_data"), nil
		},
	}

	secretService := NewSecretService(mockSecretRepo, mockEncryptor)

	ctx := context.Background()
	secret, metadata, err := secretService.GetSecret(ctx, 456, 123)

	assert.NoError(t, err)
	assert.NotNil(t, secret)
	assert.NotNil(t, metadata)
	assert.Equal(t, 456, secret.ID)
	assert.Equal(t, 123, secret.UserID)
	assert.Equal(t, []byte("decrypted_data"), secret.EncryptedData)
	assert.Equal(t, "Test Secret", metadata.Name)
	assert.Equal(t, "Test Description", metadata.Description)
}

func TestSecretService_UpdateSecret_Success(t *testing.T) {
	mockSecretRepo := &MockSecretRepository{
		GetSecretByIDFunc: func(ctx context.Context, id, userID int) (*models.Secret, error) {
			return &models.Secret{
				ID:            456,
				UserID:        123,
				Type:          models.TypeLogin,
				Name:          "Old Name",
				EncryptedData: []byte("old_encrypted_data"),
				Version:       1,
				CreatedAt:     time.Now().Add(-24 * time.Hour),
				UpdatedAt:     time.Now().Add(-1 * time.Hour),
			}, nil
		},
		UpdateSecretFunc: func(ctx context.Context, secret *models.Secret) error {
			assert.Equal(t, 456, secret.ID)
			assert.Equal(t, 123, secret.UserID)
			assert.Equal(t, []byte("new_encrypted_data"), secret.EncryptedData)
			assert.Equal(t, int32(2), secret.Version)
			return nil
		},
	}

	mockEncryptor := &MockEncryptor{
		EncryptFunc: func(data []byte) ([]byte, error) {
			return []byte("new_encrypted_data"), nil
		},
	}

	secretService := NewSecretService(mockSecretRepo, mockEncryptor)

	ctx := context.Background()
	metadata := &v1.Metadata{
		Name: "Updated Secret",
	}

	updatedSecret, err := secretService.UpdateSecret(ctx, 456, 123, "Updated Secret", []byte("new_data"), metadata, 1)

	assert.NoError(t, err)
	assert.NotNil(t, updatedSecret)
	assert.Equal(t, 456, updatedSecret.ID)
	assert.Equal(t, "Updated Secret", updatedSecret.Name)
	assert.Equal(t, int32(2), updatedSecret.Version)
}

func TestSecretService_UpdateSecret_VersionConflict(t *testing.T) {
	mockSecretRepo := &MockSecretRepository{
		GetSecretByIDFunc: func(ctx context.Context, id, userID int) (*models.Secret, error) {
			return &models.Secret{
				ID:      456,
				UserID:  123,
				Version: 2,
			}, nil
		},
	}

	mockEncryptor := &MockEncryptor{}
	secretService := NewSecretService(mockSecretRepo, mockEncryptor)

	ctx := context.Background()
	metadata := &v1.Metadata{Name: "Updated Secret"}

	updatedSecret, err := secretService.UpdateSecret(ctx, 456, 123, "Updated Secret", []byte("new_data"), metadata, 1)

	assert.Error(t, err)
	assert.Nil(t, updatedSecret)
	assert.Contains(t, err.Error(), "version conflict")
}

func TestSecretService_DeleteSecret_Success(t *testing.T) {
	mockSecretRepo := &MockSecretRepository{
		DeleteSecretFunc: func(ctx context.Context, id, userID int) error {
			assert.Equal(t, 456, id)
			assert.Equal(t, 123, userID)
			return nil
		},
	}

	mockEncryptor := &MockEncryptor{}
	secretService := NewSecretService(mockSecretRepo, mockEncryptor)

	ctx := context.Background()
	err := secretService.DeleteSecret(ctx, 456, 123)

	assert.NoError(t, err)
}

func TestSecretService_ListSecrets_Success(t *testing.T) {
	secrets := []*models.Secret{
		{
			ID:      1,
			UserID:  123,
			Type:    models.TypeLogin,
			Name:    "Secret 1",
			Version: 1,
		},
		{
			ID:      2,
			UserID:  123,
			Type:    models.TypeCard,
			Name:    "Secret 2",
			Version: 1,
		},
	}

	mockSecretRepo := &MockSecretRepository{
		ListByUserIDFunc: func(ctx context.Context, userID int) ([]*models.Secret, error) {
			assert.Equal(t, 123, userID)
			return secrets, nil
		},
	}

	mockEncryptor := &MockEncryptor{}
	secretService := NewSecretService(mockSecretRepo, mockEncryptor)

	ctx := context.Background()
	result, err := secretService.ListSecrets(ctx, 123)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, 1, result[0].ID)
	assert.Equal(t, 2, result[1].ID)
}

func TestSecretService_SyncSecrets_Success(t *testing.T) {
	serverSecrets := []*models.Secret{
		{
			ID:      1,
			UserID:  123,
			Type:    models.TypeLogin,
			Name:    "Server Secret",
			Version: 1,
		},
	}

	mockSecretRepo := &MockSecretRepository{
		ListByUserIDFunc: func(ctx context.Context, userID int) ([]*models.Secret, error) {
			assert.Equal(t, 123, userID)
			return serverSecrets, nil
		},
	}

	mockEncryptor := &MockEncryptor{}
	secretService := NewSecretService(mockSecretRepo, mockEncryptor)

	ctx := context.Background()
	clientSecrets := []*models.Secret{
		{
			ID:      2,
			UserID:  123,
			Type:    models.TypeCard,
			Name:    "Client Secret",
			Version: 1,
		},
	}

	result, err := secretService.SyncSecrets(ctx, 123, clientSecrets)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Server Secret", result[0].Name)
}
