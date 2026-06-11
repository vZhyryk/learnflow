package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type QueryRunner interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

type txKey struct{}

type PgxTransactor struct {
	pool *pgxpool.Pool
}

func NewTransactor(pool *pgxpool.Pool) *PgxTransactor {
	return &PgxTransactor{pool: pool}
}

func (t *PgxTransactor) InTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := t.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("transactor.BeginTx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err = fn(context.WithValue(ctx, txKey{}, tx)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// ExtractTx returns the active transaction from context, if any.
func ExtractTx(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	return tx, ok
}
