package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type Connection interface {
	Connection() (*sql.DB, error)
	Close(ctx context.Context) error
}

type PostgresConnection struct {
	db *sql.DB
}

func NewPostgresConnection(
	dbUser string,
	dbPassword string,
	dbHost string,
	dbPort int,
	dbName string,
	dbSSLMode string) (*PostgresConnection, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		dbHost,
		dbPort,
		dbUser,
		dbPassword,
		dbName,
		dbSSLMode,
	)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &PostgresConnection{db: db}, nil
}

func (p *PostgresConnection) Connection() (*sql.DB, error) {
	if err := p.db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	return p.db, nil
}

func (p *PostgresConnection) Close(ctx context.Context) error {
	if err := p.db.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}
	return nil
}
