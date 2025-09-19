package postgres

import (
	"context"
	"fmt"

	"go-svc-gophkeeper/internal/models"

	"github.com/jackc/pgx/v5"
)

// CreateSecret создает новую запись секрета
func (s *Store) CreateSecret(ctx context.Context, secret *models.Secret) (int, error) {
	query := `
		INSERT INTO secrets (user_id, type, name, metadata, encrypted_data, version)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	var id int
	err := s.pool.QueryRow(ctx, query,
		secret.UserID,
		secret.Type,
		secret.Name,
		secret.Metadata,
		secret.EncryptedData,
		secret.Version,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to create secret: %w", err)
	}

	return id, nil
}

// GetSecretByID возвращает секрет по ID с проверкой принадлежности пользователю
func (s *Store) GetSecretByID(ctx context.Context, id, userID int) (*models.Secret, error) {
	query := `
		SELECT id, user_id, type, name, metadata, encrypted_data, version, 
		       deleted_at, created_at, updated_at
		FROM secrets
		WHERE id = $1 AND user_id = $2
	`

	var secret models.Secret
	err := s.pool.QueryRow(ctx, query, id, userID).Scan(
		&secret.ID,
		&secret.UserID,
		&secret.Type,
		&secret.Name,
		&secret.Metadata,
		&secret.EncryptedData,
		&secret.Version,
		&secret.DeletedAt,
		&secret.CreatedAt,
		&secret.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get secret by ID: %w", err)
	}

	return &secret, nil
}

// ListByUserID возвращает все секреты пользователя
func (s *Store) ListByUserID(ctx context.Context, userID int) ([]*models.Secret, error) {
	query := `
		SELECT id, user_id, type, name, metadata, encrypted_data, version,
		       deleted_at, created_at, updated_at
		FROM secrets
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := s.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query secrets: %w", err)
	}
	defer rows.Close()

	var secrets []*models.Secret
	for rows.Next() {
		var secret models.Secret
		err := rows.Scan(
			&secret.ID,
			&secret.UserID,
			&secret.Type,
			&secret.Name,
			&secret.Metadata,
			&secret.EncryptedData,
			&secret.Version,
			&secret.DeletedAt,
			&secret.CreatedAt,
			&secret.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan secret: %w", err)
		}
		secrets = append(secrets, &secret)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return secrets, nil
}

// UpdateSecret обновляет данные секрета
func (s *Store) UpdateSecret(ctx context.Context, secret *models.Secret) error {
	query := `
		UPDATE secrets 
		SET type = $1, name = $2, metadata = $3, encrypted_data = $4, 
		    version = $5, deleted_at = $6, updated_at = CURRENT_TIMESTAMP
		WHERE id = $7 AND user_id = $8
	`

	result, err := s.pool.Exec(ctx, query,
		secret.Type,
		secret.Name,
		secret.Metadata,
		secret.EncryptedData,
		secret.Version,
		secret.DeletedAt,
		secret.ID,
		secret.UserID,
	)

	if err != nil {
		return fmt.Errorf("failed to update secret: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("secret not found")
	}

	return nil
}

// DeleteSecret помечает секрет как удаленный (soft delete)
func (s *Store) DeleteSecret(ctx context.Context, id, userID int) error {
	query := `
		UPDATE secrets 
		SET deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND user_id = $2
	`

	result, err := s.pool.Exec(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("secret not found")
	}

	return nil
}
