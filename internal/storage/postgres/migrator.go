package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// PostgresMigrator — мигратор для Postgres
type PostgresMigrator struct {
	pool           *pgxpool.Pool
	migrationsPath string
}

// NewMigrator создаёт новый мигратор
func NewMigrator(ctx context.Context, pool *pgxpool.Pool, migrationsPath string) (*PostgresMigrator, error) {
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("migrations directory does not exist: %s", migrationsPath)
	}

	goose.SetBaseFS(os.DirFS(migrationsPath))
	goose.SetTableName("goose_migrations")
	goose.SetVerbose(true)

	return &PostgresMigrator{
		pool:           pool,
		migrationsPath: migrationsPath,
	}, nil
}

// getSQLDB создаёт *sql.DB из pgxpool
func (m *PostgresMigrator) getSQLDB() (*sql.DB, error) {
	connConfig := m.pool.Config().ConnConfig.Copy()
	sqlDB := stdlib.OpenDB(*connConfig)
	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		return nil, err
	}
	return sqlDB, nil
}

// RunMigrations применяет все миграции
func (m *PostgresMigrator) RunMigrations(ctx context.Context) error {
	sqlDB, err := m.getSQLDB()
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	if version, err := goose.GetDBVersion(sqlDB); err == nil {
		log.Printf("Current DB version: %d", version)
	} else {
		log.Printf("Could not get DB version: %v", err)
	}

	if err := goose.Up(sqlDB, "."); err != nil {
		return err
	}

	log.Println("Migrations applied successfully")
	return nil
}
