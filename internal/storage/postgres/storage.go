package postgres

import (
	"context"
	"fmt"
	"time"

	"go-svc-gophkeeper/internal/storage"

	"github.com/jackc/pgx/v5/pgxpool"
)

// _ Проверка на этапе компиляции, что наша структура реализует интерфейс.
var _ storage.Storage = (*Store)(nil)

// Store реализует интерфейс storage.Storage.
type Store struct {
	// Заменяем *sqlx.DB на *pgxpool.Pool
	pool *pgxpool.Pool
}

// New - конструктор хранилища для pgxpool.
func New(dsn string) (*Store, error) {
	// Создаем конфиг пула из DSN
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DSN: %w", err)
	}

	// Настраиваем пул соединений (опционально, но рекомендуется)
	config.MaxConns = 25                     // Максимальное количество соединений в пуле
	config.MinConns = 5                      // Минимальное количество соединений в пуле
	config.MaxConnLifetime = 5 * time.Minute // Макс. время жизни соединения
	config.MaxConnIdleTime = 1 * time.Minute // Макс. время простаивания соединения

	// Создаем пул соединений с контекстом
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Проверяем подключение
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Store{pool: pool}, nil
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
