package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

type Transaction interface {
	WithTransaction(ctx context.Context, fn func(tx *sql.Tx) error) error
}

type PostgresTransaction struct {
	DB *sql.DB
}

func NewPostgresTransaction(db *sql.DB) *PostgresTransaction {
	return &PostgresTransaction{DB: db}
}

func (p *PostgresTransaction) WithTransaction(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := p.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction error: %v, rollback error: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}
