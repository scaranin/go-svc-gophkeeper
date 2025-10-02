package client

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	v1 "go-svc-gophkeeper/gen/go/v1"
	"go-svc-gophkeeper/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// MockAuthServiceClient мок для AuthServiceClient
type MockAuthServiceClient struct {
	RegisterFunc func(ctx context.Context, in *v1.RegisterRequest, opts ...grpc.CallOption) (*v1.RegisterResponse, error)
	LoginFunc    func(ctx context.Context, in *v1.LoginRequest, opts ...grpc.CallOption) (*v1.LoginResponse, error)
}

func (m *MockAuthServiceClient) Register(ctx context.Context, in *v1.RegisterRequest, opts ...grpc.CallOption) (*v1.RegisterResponse, error) {
	if m.RegisterFunc != nil {
		return m.RegisterFunc(ctx, in, opts...)
	}
	return nil, errors.New("RegisterFunc not implemented")
}

func (m *MockAuthServiceClient) Login(ctx context.Context, in *v1.LoginRequest, opts ...grpc.CallOption) (*v1.LoginResponse, error) {
	if m.LoginFunc != nil {
		return m.LoginFunc(ctx, in, opts...)
	}
	return nil, errors.New("LoginFunc not implemented")
}

// MockSecretServiceClient мок для SecretServiceClient
type MockSecretServiceClient struct {
	CreateSecretFunc func(ctx context.Context, in *v1.CreateSecretRequest, opts ...grpc.CallOption) (*v1.CreateSecretResponse, error)
	ListSecretsFunc  func(ctx context.Context, in *v1.ListSecretsRequest, opts ...grpc.CallOption) (*v1.ListSecretsResponse, error)
	GetSecretFunc    func(ctx context.Context, in *v1.GetSecretRequest, opts ...grpc.CallOption) (*v1.GetSecretResponse, error)
	UpdateSecretFunc func(ctx context.Context, in *v1.UpdateSecretRequest, opts ...grpc.CallOption) (*v1.UpdateSecretResponse, error)
	DeleteSecretFunc func(ctx context.Context, in *v1.DeleteSecretRequest, opts ...grpc.CallOption) (*v1.DeleteSecretResponse, error)
	SyncFunc         func(ctx context.Context, in *v1.SyncRequest, opts ...grpc.CallOption) (*v1.SyncResponse, error)
}

func (m *MockSecretServiceClient) CreateSecret(ctx context.Context, in *v1.CreateSecretRequest, opts ...grpc.CallOption) (*v1.CreateSecretResponse, error) {
	if m.CreateSecretFunc != nil {
		return m.CreateSecretFunc(ctx, in, opts...)
	}
	return nil, errors.New("CreateSecretFunc not implemented")
}

func (m *MockSecretServiceClient) ListSecrets(ctx context.Context, in *v1.ListSecretsRequest, opts ...grpc.CallOption) (*v1.ListSecretsResponse, error) {
	if m.ListSecretsFunc != nil {
		return m.ListSecretsFunc(ctx, in, opts...)
	}
	return nil, errors.New("ListSecretsFunc not implemented")
}

func (m *MockSecretServiceClient) GetSecret(ctx context.Context, in *v1.GetSecretRequest, opts ...grpc.CallOption) (*v1.GetSecretResponse, error) {
	if m.GetSecretFunc != nil {
		return m.GetSecretFunc(ctx, in, opts...)
	}
	return nil, errors.New("GetSecretFunc not implemented")
}

func (m *MockSecretServiceClient) UpdateSecret(ctx context.Context, in *v1.UpdateSecretRequest, opts ...grpc.CallOption) (*v1.UpdateSecretResponse, error) {
	if m.UpdateSecretFunc != nil {
		return m.UpdateSecretFunc(ctx, in, opts...)
	}
	return nil, errors.New("UpdateSecretFunc not implemented")
}

func (m *MockSecretServiceClient) DeleteSecret(ctx context.Context, in *v1.DeleteSecretRequest, opts ...grpc.CallOption) (*v1.DeleteSecretResponse, error) {
	if m.DeleteSecretFunc != nil {
		return m.DeleteSecretFunc(ctx, in, opts...)
	}
	return nil, errors.New("DeleteSecretFunc not implemented")
}

func (m *MockSecretServiceClient) Sync(ctx context.Context, in *v1.SyncRequest, opts ...grpc.CallOption) (*v1.SyncResponse, error) {
	if m.SyncFunc != nil {
		return m.SyncFunc(ctx, in, opts...)
	}
	return nil, errors.New("SyncFunc not implemented")
}

func createTestConfig(t *testing.T) *config.ClientConfig {
	tempDir := t.TempDir()

	return &config.ClientConfig{
		ServerURL:  "localhost:50051",
		Timeout:    30 * time.Second,
		ConfigPath: "",
		DataDir:    tempDir,
	}
}

func createTestApp(t *testing.T, authClient v1.AuthServiceClient, secretClient v1.SecretServiceClient) *App {
	cfg := createTestConfig(t)

	app := &App{
		config:       cfg,
		authClient:   authClient,
		secretClient: secretClient,
	}

	err := app.ensureDataDir()
	require.NoError(t, err)

	return app
}

// Tests
func TestApp_Register_Success(t *testing.T) {
	mockAuth := &MockAuthServiceClient{
		RegisterFunc: func(ctx context.Context, in *v1.RegisterRequest, opts ...grpc.CallOption) (*v1.RegisterResponse, error) {
			assert.Equal(t, "testuser", in.Login)
			assert.Equal(t, "testpass", in.Password)
			return &v1.RegisterResponse{
				UserId:      "user123",
				AccessToken: "token123",
			}, nil
		},
	}

	app := createTestApp(t, mockAuth, &MockSecretServiceClient{})
	ctx := context.Background()

	err := app.Register(ctx, "testuser", "testpass")

	assert.NoError(t, err)
	assert.Equal(t, "token123", app.GetToken())
}

func TestApp_Register_Error(t *testing.T) {
	mockAuth := &MockAuthServiceClient{
		RegisterFunc: func(ctx context.Context, in *v1.RegisterRequest, opts ...grpc.CallOption) (*v1.RegisterResponse, error) {
			return nil, errors.New("registration failed")
		},
	}

	app := createTestApp(t, mockAuth, &MockSecretServiceClient{})
	ctx := context.Background()

	err := app.Register(ctx, "testuser", "testpass")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "registration failed")
}

func TestApp_Login_Success(t *testing.T) {
	mockAuth := &MockAuthServiceClient{
		LoginFunc: func(ctx context.Context, in *v1.LoginRequest, opts ...grpc.CallOption) (*v1.LoginResponse, error) {
			assert.Equal(t, "testuser", in.Login)
			assert.Equal(t, "testpass", in.Password)
			return &v1.LoginResponse{
				AccessToken: "token456",
				UserId:      "user456",
				ExpiresAt:   timestamppb.New(time.Now().Add(24 * time.Hour)),
			}, nil
		},
	}

	app := createTestApp(t, mockAuth, &MockSecretServiceClient{})
	ctx := context.Background()

	err := app.Login(ctx, "testuser", "testpass")

	assert.NoError(t, err)
	assert.Equal(t, "token456", app.GetToken())
}

func TestApp_CreateSecret_Success(t *testing.T) {
	mockSecret := &MockSecretServiceClient{
		CreateSecretFunc: func(ctx context.Context, in *v1.CreateSecretRequest, opts ...grpc.CallOption) (*v1.CreateSecretResponse, error) {
			assert.Equal(t, v1.SecretType_SECRET_TYPE_LOGIN, in.Type)
			assert.Equal(t, "Test Secret", in.Name)
			assert.Equal(t, []byte("encrypted_data"), in.EncryptedData)

			return &v1.CreateSecretResponse{
				SecretId: "secret123",
				Version:  1,
			}, nil
		},
	}

	app := createTestApp(t, &MockAuthServiceClient{}, mockSecret)
	app.SetToken("test-token")

	ctx := context.Background()
	metadata := &v1.Metadata{
		Name:        "Test Secret",
		Description: "Test Description",
		Website:     "https://example.com",
		Tags:        "test",
	}

	err := app.CreateSecret(ctx, v1.SecretType_SECRET_TYPE_LOGIN, "Test Secret", metadata, []byte("encrypted_data"))

	assert.NoError(t, err)
}

func TestApp_ListSecrets_Success(t *testing.T) {
	mockSecret := &MockSecretServiceClient{
		ListSecretsFunc: func(ctx context.Context, in *v1.ListSecretsRequest, opts ...grpc.CallOption) (*v1.ListSecretsResponse, error) {
			assert.False(t, in.IncludeDeleted)

			return &v1.ListSecretsResponse{
				Secrets: []*v1.SecretInfo{
					{
						Id:      "secret1",
						Type:    v1.SecretType_SECRET_TYPE_LOGIN,
						Name:    "Google Account",
						Version: 1,
					},
					{
						Id:      "secret2",
						Type:    v1.SecretType_SECRET_TYPE_CARD,
						Name:    "Credit Card",
						Version: 1,
					},
				},
			}, nil
		},
	}

	app := createTestApp(t, &MockAuthServiceClient{}, mockSecret)
	app.SetToken("test-token")

	ctx := context.Background()

	err := app.ListSecrets(ctx, false)

	assert.NoError(t, err)
}

func TestApp_GetSecret_Success(t *testing.T) {
	mockSecret := &MockSecretServiceClient{
		GetSecretFunc: func(ctx context.Context, in *v1.GetSecretRequest, opts ...grpc.CallOption) (*v1.GetSecretResponse, error) {
			assert.Equal(t, "secret123", in.SecretId)

			return &v1.GetSecretResponse{
				Info: &v1.SecretInfo{
					Id:      "secret123",
					Type:    v1.SecretType_SECRET_TYPE_LOGIN,
					Name:    "Google Account",
					Version: 1,
					Metadata: &v1.Metadata{
						Name:        "Google Account",
						Description: "Gmail password",
						Website:     "https://gmail.com",
						Tags:        "email",
					},
					CreatedAt: timestamppb.New(time.Now().Add(-24 * time.Hour)),
					UpdatedAt: timestamppb.New(time.Now()),
				},
				EncryptedData: []byte("encrypted_login_data"),
			}, nil
		},
	}

	app := createTestApp(t, &MockAuthServiceClient{}, mockSecret)
	app.SetToken("test-token")

	ctx := context.Background()

	err := app.GetSecret(ctx, "secret123")

	assert.NoError(t, err)
}

func TestApp_UpdateSecret_Success(t *testing.T) {
	mockSecret := &MockSecretServiceClient{
		UpdateSecretFunc: func(ctx context.Context, in *v1.UpdateSecretRequest, opts ...grpc.CallOption) (*v1.UpdateSecretResponse, error) {
			assert.Equal(t, "secret123", in.SecretId)
			assert.Equal(t, []byte("new_encrypted_data"), in.EncryptedData)

			return &v1.UpdateSecretResponse{
				NewVersion: 2,
				UpdatedAt:  timestamppb.New(time.Now()),
			}, nil
		},
	}

	app := createTestApp(t, &MockAuthServiceClient{}, mockSecret)
	app.SetToken("test-token")

	ctx := context.Background()

	err := app.UpdateSecret(ctx, "secret123", []byte("new_encrypted_data"))

	assert.NoError(t, err)
}

func TestApp_DeleteSecret_Success(t *testing.T) {
	mockSecret := &MockSecretServiceClient{
		DeleteSecretFunc: func(ctx context.Context, in *v1.DeleteSecretRequest, opts ...grpc.CallOption) (*v1.DeleteSecretResponse, error) {
			assert.Equal(t, "secret123", in.SecretId)

			return &v1.DeleteSecretResponse{
				Success:   true,
				DeletedAt: timestamppb.New(time.Now()),
			}, nil
		},
	}

	app := createTestApp(t, &MockAuthServiceClient{}, mockSecret)
	app.SetToken("test-token")

	ctx := context.Background()

	err := app.DeleteSecret(ctx, "secret123")

	assert.NoError(t, err)
}

func TestApp_Sync_Success(t *testing.T) {
	mockSecret := &MockSecretServiceClient{
		SyncFunc: func(ctx context.Context, in *v1.SyncRequest, opts ...grpc.CallOption) (*v1.SyncResponse, error) {
			assert.Nil(t, in.LastSync) // Testing with nil lastSync

			return &v1.SyncResponse{
				Secrets: []*v1.SecretInfo{
					{
						Id:      "secret1",
						Type:    v1.SecretType_SECRET_TYPE_LOGIN,
						Name:    "Synced Secret",
						Version: 1,
					},
				},
				SyncState: &v1.SyncState{
					TotalSecrets: 1,
					LastSync:     timestamppb.New(time.Now()),
				},
			}, nil
		},
	}

	app := createTestApp(t, &MockAuthServiceClient{}, mockSecret)
	app.SetToken("test-token")

	ctx := context.Background()

	err := app.Sync(ctx, nil)

	assert.NoError(t, err)
}

func TestApp_Sync_WithLastSync(t *testing.T) {
	lastSyncTime := time.Now().Add(-1 * time.Hour)
	lastSync := timestamppb.New(lastSyncTime)

	mockSecret := &MockSecretServiceClient{
		SyncFunc: func(ctx context.Context, in *v1.SyncRequest, opts ...grpc.CallOption) (*v1.SyncResponse, error) {
			assert.NotNil(t, in.LastSync)
			assert.Equal(t, lastSyncTime.Unix(), in.LastSync.AsTime().Unix())

			return &v1.SyncResponse{
				Secrets: []*v1.SecretInfo{},
				SyncState: &v1.SyncState{
					TotalSecrets: 0,
					LastSync:     timestamppb.New(time.Now()),
				},
			}, nil
		},
	}

	app := createTestApp(t, &MockAuthServiceClient{}, mockSecret)
	app.SetToken("test-token")

	ctx := context.Background()

	err := app.Sync(ctx, lastSync)

	assert.NoError(t, err)
}

func TestApp_IsAuthenticated(t *testing.T) {
	app := createTestApp(t, &MockAuthServiceClient{}, &MockSecretServiceClient{})

	// Initially not authenticated
	assert.False(t, app.IsAuthenticated())

	// After setting token
	app.SetToken("test-token")
	assert.True(t, app.IsAuthenticated())

	// After clearing token
	app.SetToken("")
	assert.False(t, app.IsAuthenticated())
}

func TestApp_EnsureDataDir(t *testing.T) {
	tempDir := t.TempDir()
	testDir := filepath.Join(tempDir, "test-data")

	cfg := &config.ClientConfig{
		DataDir: testDir,
	}

	app := &App{config: cfg}

	// Directory should not exist initially
	_, err := os.Stat(testDir)
	assert.True(t, os.IsNotExist(err))

	// After ensureDataDir, directory should exist
	err = app.ensureDataDir()
	assert.NoError(t, err)

	_, err = os.Stat(testDir)
	assert.NoError(t, err)

	// Check permissions
	info, err := os.Stat(testDir)
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0700), info.Mode().Perm())
}

func TestApp_EnsureDataDir_WithTilde(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	testDir := filepath.Join(homeDir, "test-gophkeeper")
	defer os.RemoveAll(testDir) // Cleanup

	cfg := &config.ClientConfig{
		DataDir: "~/test-gophkeeper",
	}

	app := &App{config: cfg}

	err = app.ensureDataDir()
	assert.NoError(t, err)

	// Check that directory was created in home directory
	_, err = os.Stat(testDir)
	assert.NoError(t, err)
}

func TestApp_CreateContext(t *testing.T) {
	app := createTestApp(t, &MockAuthServiceClient{}, &MockSecretServiceClient{})

	ctx := context.Background()
	newCtx, cancel := app.createContext(ctx)

	assert.NotNil(t, newCtx)
	assert.NotNil(t, cancel)

	// Ensure cancel works
	cancel()

	select {
	case <-newCtx.Done():
		// Context should be cancelled
	default:
		t.Error("Context should be cancelled after calling cancel")
	}
}

func TestApp_CreateAuthContext(t *testing.T) {
	app := createTestApp(t, &MockAuthServiceClient{}, &MockSecretServiceClient{})

	ctx := context.Background()

	// Without token
	authCtx := app.createAuthContext(ctx)
	assert.Equal(t, ctx, authCtx)

	// With token
	app.SetToken("test-token")
	authCtx = app.createAuthContext(ctx)
	assert.NotEqual(t, ctx, authCtx)
}
