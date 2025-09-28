package postgres

import (
	"context"
	"fmt"

	"go-svc-gophkeeper/internal/models"

	"github.com/jackc/pgx/v5"
)

// CreateSecret создает новую запись секрета
func (s *Store) CreateSecret(ctx context.Context, secret *models.Secret) (int, error) {
	query := `INSERT 
	            INTO secrets 
				   ( user_id
				   , type
				   , name
				   , metadata
				   , encrypted_data
				   , version
				   )
			  VALUES 
			       ( @user_id
				   , @type
				   , @name
				   , @metadata
				   , @encrypted_data
				   , @version
				   )
		   RETURNING id`

	args := pgx.NamedArgs{
		"user_id":        secret.UserID,
		"type":           secret.Type,
		"name":           secret.Name,
		"metadata":       secret.Metadata,
		"encrypted_data": secret.EncryptedData,
		"version":        secret.Version,
	}

	var id int
	err := s.pool.QueryRow(ctx, query, args).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to create secret: %w", err)
	}

	return id, nil
}

// GetSecretByID возвращает секрет по ID с проверкой принадлежности пользователю
func (s *Store) GetSecretByID(ctx context.Context, id, userID int) (*models.Secret, error) {
	query := `SELECT id, user_id, type, name, metadata, encrypted_data, version, deleted_at, created_at, updated_at
		        FROM secrets
		       WHERE id 	 = @id
			     AND user_id = @user_id`

	args := pgx.NamedArgs{
		"id":      id,
		"user_id": userID,
	}

	var secret models.Secret
	err := s.pool.QueryRow(ctx, query, args).Scan(
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
	query := `SELECT id, user_id, type, name, metadata, encrypted_data, version, deleted_at, created_at, updated_at
		        FROM secrets
		       WHERE user_id = @user_id`

	args := pgx.NamedArgs{
		"user_id": userID,
	}

	rows, err := s.pool.Query(ctx, query, args)
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
	query := `UPDATE secrets 
		         SET type           = @type
				   , name           = @name
				   , metadata       = @metadata
				   , encrypted_data = @encrypted_data
				   , version        = @version
				   , deleted_at     = @deleted_at
				   , updated_at     = CURRENT_TIMESTAMP
		       WHERE id      = @id
			     AND user_id = @user_id`

	args := pgx.NamedArgs{
		"type":           secret.Type,
		"name":           secret.Name,
		"metadata":       secret.Metadata,
		"encrypted_data": secret.EncryptedData,
		"version":        secret.Version,
		"deleted_at":     secret.DeletedAt,
		"id":             secret.ID,
		"user_id":        secret.UserID,
	}

	result, err := s.pool.Exec(ctx, query, args)

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
	query := `UPDATE secrets 
		         SET deleted_at = CURRENT_TIMESTAMP
				   , updated_at = CURRENT_TIMESTAMP
		       WHERE id 	 = @id 
			     AND user_id = @user_id`

	args := pgx.NamedArgs{
		"id":      id,
		"user_id": userID,
	}

	result, err := s.pool.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("secret not found")
	}

	return nil
}
