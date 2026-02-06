package database

import (
	"database/sql"
	"fmt"

	"apidian-go/internal/config"

	_ "github.com/lib/pq"
)

type Database struct {
	DB *sql.DB
}

func NewPostgresConnection(cfg *config.DatabaseConfig) (*Database, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.DBName,
		cfg.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	return &Database{DB: db}, nil
}

func (d *Database) Close() error {
	return d.DB.Close()
}
