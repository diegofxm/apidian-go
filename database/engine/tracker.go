package engine

import (
	"database/sql"
	"fmt"
	"time"
)

// Tracker maneja el tracking de migraciones en schema_migrations
type Tracker struct {
	db *sql.DB
}

// NewTracker crea un nuevo tracker
func NewTracker(db *sql.DB) *Tracker {
	return &Tracker{db: db}
}

// EnsureMigrationsTable crea la tabla schema_migrations si no existe
func (t *Tracker) EnsureMigrationsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL UNIQUE,
			batch INTEGER NOT NULL,
			executed_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_schema_migrations_batch ON schema_migrations(batch);
	`
	_, err := t.db.Exec(query)
	return err
}

// RecordMigration registra una migración ejecutada
func (t *Tracker) RecordMigration(name string, batch int) error {
	query := `INSERT INTO schema_migrations (name, batch, executed_at) VALUES ($1, $2, $3)`
	_, err := t.db.Exec(query, name, batch, time.Now())
	return err
}

// RemoveMigration elimina un registro de migración
func (t *Tracker) RemoveMigration(name string) error {
	query := `DELETE FROM schema_migrations WHERE name = $1`
	_, err := t.db.Exec(query, name)
	return err
}

// GetExecutedMigrations obtiene todas las migraciones ejecutadas
func (t *Tracker) GetExecutedMigrations() ([]MigrationRecord, error) {
	query := `SELECT id, name, batch, executed_at FROM schema_migrations ORDER BY id ASC`
	rows, err := t.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var migrations []MigrationRecord
	for rows.Next() {
		var m MigrationRecord
		var executedAt time.Time
		if err := rows.Scan(&m.ID, &m.Name, &m.Batch, &executedAt); err != nil {
			return nil, err
		}
		m.ExecutedAt = executedAt.Format("2006-01-02 15:04:05")
		migrations = append(migrations, m)
	}

	return migrations, nil
}

// GetNextBatch obtiene el siguiente número de batch
func (t *Tracker) GetNextBatch() (int, error) {
	var maxBatch sql.NullInt64
	query := `SELECT MAX(batch) FROM schema_migrations`
	err := t.db.QueryRow(query).Scan(&maxBatch)
	if err != nil {
		return 0, err
	}

	if maxBatch.Valid {
		return int(maxBatch.Int64) + 1, nil
	}
	return 1, nil
}

// IsMigrationExecuted verifica si una migración ya fue ejecutada
func (t *Tracker) IsMigrationExecuted(name string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM schema_migrations WHERE name = $1`
	err := t.db.QueryRow(query, name).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetLastBatch obtiene el último batch ejecutado
func (t *Tracker) GetLastBatch() (int, error) {
	var maxBatch sql.NullInt64
	query := `SELECT MAX(batch) FROM schema_migrations`
	err := t.db.QueryRow(query).Scan(&maxBatch)
	if err != nil {
		return 0, err
	}

	if maxBatch.Valid {
		return int(maxBatch.Int64), nil
	}
	return 0, fmt.Errorf("no migrations found")
}

// DropAllTables elimina todas las tablas de la base de datos
func (t *Tracker) DropAllTables() error {
	query := `
		DO $$ 
		DECLARE
			r RECORD;
		BEGIN
			FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public') LOOP
				EXECUTE 'DROP TABLE IF EXISTS ' || quote_ident(r.tablename) || ' CASCADE';
			END LOOP;
		END $$;
	`
	_, err := t.db.Exec(query)
	return err
}
