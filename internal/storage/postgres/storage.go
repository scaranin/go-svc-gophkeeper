package postgres

import (
	"context"
	"fmt"
	"time"

	"go-svc-gophkeeper/internal/storage"

	"github.com/jackc/pgx/v5/pgxpool"
)

var _ storage.Storage = (*Store)(nil)

type Store struct {
	pool     *pgxpool.Pool
	migrator *PostgresMigrator
}

// New инициализирует хранилище Postgres
func New(dsn, migrationsPath string) (*Store, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := NewPool(ctx, dsn)
	if err != nil {
		return nil, err
	}

	migrator, err := NewMigrator(ctx, pool, migrationsPath)
	if err != nil {
		return nil, err
	}

	store := &Store{
		pool:     pool,
		migrator: migrator,
	}

	if err := store.Migrate(ctx); err != nil {
		store.Close()
		return nil, err
	}

	return store, nil
}

// NewPool инициализирует пул соединений Postgres
func NewPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DSN: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}

// Migrate запускает миграции БД
func (s *Store) Migrate(ctx context.Context) error {
	return s.migrator.RunMigrations(ctx)
}

// User возвращает реализацию интерфейса UserRepository для Postgres.
func (s *Store) User() storage.UserRepository {
	return s
}

// Secret возвращает реализацию интерфейса SecretRepository для Postgres.
func (s *Store) Secret() storage.SecretRepository {
	return s
}

// Close закрывает пул соединений с базой данных.
func (s *Store) Close() {
	s.pool.Close()
}

// Ping проверяет подключение к БД
func (s *Store) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}
