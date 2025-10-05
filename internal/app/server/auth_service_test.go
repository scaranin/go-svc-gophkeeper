package server

import (
	"context"
	"testing"
	"time"

	"go-svc-gophkeeper/internal/auth"
	"go-svc-gophkeeper/internal/errors"
	"go-svc-gophkeeper/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockUserRepository мок для репозитория пользователей
type MockUserRepository struct {
	CreateUserFunc     func(ctx context.Context, user *models.User) (int, error)
	GetUserByLoginFunc func(ctx context.Context, login string) (*models.User, error)
	GetUserByIDFunc    func(ctx context.Context, id int) (*models.User, error)
	UpdateUserFunc     func(ctx context.Context, user *models.User) error
	DeleteUserFunc     func(ctx context.Context, id int) error
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *models.User) (int, error) {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(ctx, user)
	}
	return 0, nil
}

func (m *MockUserRepository) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	if m.GetUserByLoginFunc != nil {
		return m.GetUserByLoginFunc(ctx, login)
	}
	return nil, nil
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	if m.GetUserByIDFunc != nil {
		return m.GetUserByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockUserRepository) UpdateUser(ctx context.Context, user *models.User) error {
	if m.UpdateUserFunc != nil {
		return m.UpdateUserFunc(ctx, user)
	}
	return nil
}

func (m *MockUserRepository) DeleteUser(ctx context.Context, id int) error {
	if m.DeleteUserFunc != nil {
		return m.DeleteUserFunc(ctx, id)
	}
	return nil
}

func TestAuthService_Register_Success(t *testing.T) {
	mockUserRepo := &MockUserRepository{
		CreateUserFunc: func(ctx context.Context, user *models.User) (int, error) {
			assert.Equal(t, "testuser", user.Login)
			assert.NotEmpty(t, user.PasswordHash)
			return 123, nil
		},
		GetUserByLoginFunc: func(ctx context.Context, login string) (*models.User, error) {
			return nil, nil
		},
	}

	jwtManager, err := auth.NewJWTManager("test-secret-key", 24*time.Hour)
	authService := NewAuthService(mockUserRepo, jwtManager)

	ctx := context.Background()
	user, token, err := authService.Register(ctx, "testuser", "testpassword123")

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.NotEmpty(t, token)
	assert.Equal(t, 123, user.ID)
	assert.Equal(t, "testuser", user.Login)
}

func TestAuthService_Register_UserAlreadyExists(t *testing.T) {
	mockUserRepo := &MockUserRepository{
		GetUserByLoginFunc: func(ctx context.Context, login string) (*models.User, error) {
			return &models.User{
				ID:    123,
				Login: "testuser",
			}, nil
		},
	}

	jwtManager, err := auth.NewJWTManager("test-secret-key", 24*time.Hour)
	authService := NewAuthService(mockUserRepo, jwtManager)

	ctx := context.Background()
	user, token, err := authService.Register(ctx, "testuser", "testpassword123")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Empty(t, token)
	assert.Equal(t, errors.ErrAlreadyExists, err)
}

func TestAuthService_Register_ShortPassword(t *testing.T) {
	mockUserRepo := &MockUserRepository{
		GetUserByLoginFunc: func(ctx context.Context, login string) (*models.User, error) {
			return nil, nil
		},
	}

	jwtManager, err := auth.NewJWTManager("test-secret-key", 24*time.Hour)
	authService := NewAuthService(mockUserRepo, jwtManager)

	ctx := context.Background()
	user, token, err := authService.Register(ctx, "testuser", "short")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Empty(t, token)
	assert.Equal(t, errors.ErrValidation, err)
}

func TestAuthService_Register_ShortLogin(t *testing.T) {
	mockUserRepo := &MockUserRepository{
		GetUserByLoginFunc: func(ctx context.Context, login string) (*models.User, error) {
			return nil, nil
		},
	}

	jwtManager, err := auth.NewJWTManager("test-secret-key", 24*time.Hour)
	authService := NewAuthService(mockUserRepo, jwtManager)

	ctx := context.Background()
	user, token, err := authService.Register(ctx, "ab", "testpassword123")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Empty(t, token)
	assert.Equal(t, errors.ErrValidation, err)
}

func TestAuthService_Login_Success(t *testing.T) {
	mockUserRepo := &MockUserRepository{
		GetUserByLoginFunc: func(ctx context.Context, login string) (*models.User, error) {
			return &models.User{
				ID:           123,
				Login:        "testuser",
				PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMye.KB1T7Qk7p2V3YpYZV7WQWJY.b5sJQW",
			}, nil
		},
	}

	jwtManager, err := auth.NewJWTManager("test-secret-key", 24*time.Hour)
	authService := NewAuthService(mockUserRepo, jwtManager)

	ctx := context.Background()
	user, token, err := authService.Login(ctx, "testuser", "testpassword123")

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.NotEmpty(t, token)
	assert.Equal(t, 123, user.ID)
	assert.Equal(t, "testuser", user.Login)
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	mockUserRepo := &MockUserRepository{
		GetUserByLoginFunc: func(ctx context.Context, login string) (*models.User, error) {
			return &models.User{
				ID:           123,
				Login:        "testuser",
				PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMye.KB1T7Qk7p2V3YpYZV7WQWJY.b5sJQW",
			}, nil
		},
	}

	jwtManager, err := auth.NewJWTManager("test-secret-key", 24*time.Hour)
	authService := NewAuthService(mockUserRepo, jwtManager)

	ctx := context.Background()
	user, token, err := authService.Login(ctx, "testuser", "wrongpassword")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Empty(t, token)
	assert.Equal(t, errors.ErrInvalidCredentials, err)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	mockUserRepo := &MockUserRepository{
		GetUserByLoginFunc: func(ctx context.Context, login string) (*models.User, error) {
			return nil, nil
		},
	}

	jwtManager, err := auth.NewJWTManager("test-secret-key", 24*time.Hour)
	authService := NewAuthService(mockUserRepo, jwtManager)

	ctx := context.Background()
	user, token, err := authService.Login(ctx, "nonexistent", "password")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Empty(t, token)
	assert.Equal(t, errors.ErrInvalidCredentials, err)
}

func TestAuthService_Login_EmptyCredentials(t *testing.T) {
	mockUserRepo := &MockUserRepository{}
	jwtManager, err := auth.NewJWTManager("test-secret-key", 24*time.Hour)
	authService := NewAuthService(mockUserRepo, jwtManager)

	ctx := context.Background()

	user, token, err := authService.Login(ctx, "", "password")
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Empty(t, token)
	assert.Equal(t, errors.ErrCredentialsRequired, err)

	user, token, err = authService.Login(ctx, "testuser", "")
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Empty(t, token)
	assert.Equal(t, errors.ErrCredentialsRequired, err)
}

func TestAuthService_ValidateToken_Success(t *testing.T) {
	mockUserRepo := &MockUserRepository{}
	jwtManager, err := auth.NewJWTManager("test-secret-key", 24*time.Hour)
	authService := NewAuthService(mockUserRepo, jwtManager)

	userID := 123
	token, err := jwtManager.GenerateToken(userID)
	require.NoError(t, err)

	validatedUserID, err := authService.ValidateToken(token)

	assert.NoError(t, err)
	assert.Equal(t, userID, validatedUserID)
}

func TestAuthService_ValidateToken_Invalid(t *testing.T) {
	mockUserRepo := &MockUserRepository{}
	jwtManager, err := auth.NewJWTManager("test-secret-key", 24*time.Hour)
	authService := NewAuthService(mockUserRepo, jwtManager)

	userID, err := authService.ValidateToken("invalid-token")

	assert.Error(t, err)
	assert.Equal(t, 0, userID)
}
