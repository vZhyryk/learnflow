package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// QueryRunner abstracts pgxpool.Pool and pgx.Tx for use in repositories.
type QueryRunner interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

type txKey struct{}

// PgxTransactor implements domain.Transactor using a pgxpool connection pool.
type PgxTransactor struct {
	pool *pgxpool.Pool
}

// NewTransactor returns a new PgxTransactor backed by the given pool.
func NewTransactor(pool *pgxpool.Pool) *PgxTransactor {
	return &PgxTransactor{pool: pool}
}

// InTransaction runs fn within a database transaction, committing on success and rolling back on error.
func (t *PgxTransactor) InTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := t.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("transactor.BeginTx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }() //nolint:errcheck // Rollback is a no-op after Commit; error intentionally ignored

	if err := fn(context.WithValue(ctx, txKey{}, tx)); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("transactor.Commit: %w", err)
	}
	return nil
}

// ExtractTx returns the active transaction from context, if any.
func ExtractTx(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	return tx, ok
}

// FallbackQueryRunner returns the active transaction from ctx (set by InTransaction) if
// present, otherwise fallback. Repositories/writers call this so the same method works
// both inside and outside a transaction without the caller having to know which.
func FallbackQueryRunner(ctx context.Context, fallback QueryRunner) QueryRunner {
	if tx, ok := ExtractTx(ctx); ok {
		return tx
	}
	return fallback
}
