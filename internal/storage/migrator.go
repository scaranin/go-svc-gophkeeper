package storage

import "context"

// Migrator интерфейс для управления миграциями БД
type Migrator interface {
	RunMigrations(ctx context.Context) error

	CheckSchema(ctx context.Context) error
}
