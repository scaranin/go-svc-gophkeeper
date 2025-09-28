package postgres

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresMigrator обертка для миграции Postgres
type PostgresMigrator struct {
	pool           *pgxpool.Pool
	migrationsPath string
}

// RunMigrations запускает миграцию при отсутствии таблиц в БД
func (m *PostgresMigrator) RunMigrations(ctx context.Context) error {
	requiredTables := []string{"users", "secrets"}
	allTablesExist := true

	query := `SELECT 
		      EXISTS ( SELECT 1
			             FROM information_schema.tables 
			            WHERE table_schema = 'public' 
			              AND table_name = $1
					 )`

	for _, table := range requiredTables {
		var exists bool
		err := m.pool.QueryRow(ctx, query, table).Scan(&exists)

		if err != nil {
			return fmt.Errorf("failed to check table %s: %w", table, err)
		}

		if !exists {
			allTablesExist = false
			log.Printf("Table %s does not exist, will apply migrations", table)
			break
		}
	}

	if allTablesExist {
		log.Println("All required tables exist, skipping migrations")
		return nil
	}

	log.Println("Applying database migrations...")
	return m.applyAllMigrations(ctx)
}

func (m *PostgresMigrator) applyAllMigrations(ctx context.Context) error {
	files, err := os.ReadDir(m.migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".up.sql") {
			log.Printf("Applying migration: %s", file.Name())

			content, err := os.ReadFile(filepath.Join(m.migrationsPath, file.Name()))
			if err != nil {
				return fmt.Errorf("failed to read migration file %s: %w", file.Name(), err)
			}

			if _, err := m.pool.Exec(ctx, string(content)); err != nil {
				return fmt.Errorf("failed to apply migration %s: %w", file.Name(), err)
			}

			log.Printf("Migration %s applied successfully", file.Name())
		}
	}

	log.Println("All migrations applied successfully")
	return nil
}

// CheckSchema проверяет таблицы на схеме
func (m *PostgresMigrator) CheckSchema(ctx context.Context) error {
	requiredTables := []string{"users", "secrets"}

	query := `SELECT 
		      	  EXISTS ( SELECT 1
			      	         FROM information_schema.tables 
			        		WHERE table_schema = 'public' 
			              	  AND table_name   = $1
					 	 )`

	for _, table := range requiredTables {
		var exists bool
		err := m.pool.QueryRow(ctx, query, table).Scan(&exists)

		if err != nil {
			return fmt.Errorf("failed to check table %s: %w", table, err)
		}

		if !exists {
			return fmt.Errorf("required table %s does not exist", table)
		}
	}

	log.Println("Database schema verified")
	return nil
}
