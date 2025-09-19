package postgres

import (
	"context"
	"fmt"

	"go-svc-gophkeeper/internal/models"

	"github.com/jackc/pgx/v5"
)

// CreateUser создает нового пользователя
func (s *Store) CreateUser(ctx context.Context, user *models.User) (int, error) {
	query := `
		INSERT INTO users (login, password_hash)
		VALUES ($1, $2)
		RETURNING id
	`

	var id int
	err := s.pool.QueryRow(ctx, query, user.Login, user.PasswordHash).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	return id, nil
}

// GetUserByID возвращает пользователя по ID
func (s *Store) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	query := `
		SELECT id, login, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user models.User
	err := s.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return &user, nil
}

// GetUserByLogin возвращает пользователя по логину
func (s *Store) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	query := `
		SELECT id, login, password_hash, created_at, updated_at
		FROM users
		WHERE login = $1
	`

	var user models.User
	err := s.pool.QueryRow(ctx, query, login).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by login: %w", err)
	}

	return &user, nil
}

// UpdateUser обновляет данные пользователя
func (s *Store) UpdateUser(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users 
		SET login = $1, password_hash = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
	`

	result, err := s.pool.Exec(ctx, query, user.Login, user.PasswordHash, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// DeleteUser удаляет пользователя
func (s *Store) DeleteUser(ctx context.Context, id int) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := s.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}
