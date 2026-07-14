//go:build integration

package db_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"learnflow_backend/internal/infrastructure/db"
	"learnflow_backend/internal/shared/testutil"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	. "github.com/smartystreets/goconvey/convey"
)

// newScratchTable creates a uniquely-named throwaway table for exercising InTransaction's commit/rollback behavior
func newScratchTable(t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()

	name := fmt.Sprintf("transactor_test_scratch_%d", time.Now().UnixNano())
	if _, err := pool.Exec(context.Background(), fmt.Sprintf("CREATE TABLE %s (id text PRIMARY KEY)", name)); err != nil {
		t.Fatalf("newScratchTable: create: %v", err)
	}
	t.Cleanup(func() {
		//nolint:errcheck // best-effort cleanup of a throwaway test table; error intentionally ignored
		pool.Exec(context.Background(), fmt.Sprintf("DROP TABLE IF EXISTS %s", name))
	})

	return name
}

// newDeferredUniqueScratchTable creates a throwaway table whose UNIQUE
// constraint is DEFERRABLE INITIALLY DEFERRED — a duplicate insert succeeds
// immediately and only fails when the constraint is checked at COMMIT time.
// This is the only way to force tx.Commit itself to fail (as opposed to
// BeginTx or an Exec inside fn), since Postgres validates immediate
// constraints eagerly on Exec, not at commit.
func newDeferredUniqueScratchTable(t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()

	name := fmt.Sprintf("transactor_test_deferred_scratch_%d", time.Now().UnixNano())
	createSQL := fmt.Sprintf(
		"CREATE TABLE %s (val text, CONSTRAINT %s_unique UNIQUE (val) DEFERRABLE INITIALLY DEFERRED)",
		name, name,
	)
	if _, err := pool.Exec(context.Background(), createSQL); err != nil {
		t.Fatalf("newDeferredUniqueScratchTable: create: %v", err)
	}
	t.Cleanup(func() {
		//nolint:errcheck // best-effort cleanup of a throwaway test table; error intentionally ignored
		pool.Exec(context.Background(), fmt.Sprintf("DROP TABLE IF EXISTS %s", name))
	})

	return name
}

// assertCommitFailureWrapped verifies that a constraint violation surfacing
// only at COMMIT time (not at BeginTx, not at any Exec inside fn) is wrapped
// with "transactor.Commit", and that the rows fn wrote do not survive.
func assertCommitFailureWrapped(t *testing.T, pool *pgxpool.Pool, transactor *db.PgxTransactor) {
	t.Helper()
	table := newDeferredUniqueScratchTable(t, pool)

	err := transactor.InTransaction(context.Background(), func(ctx context.Context) error {
		qr := db.FallbackQueryRunner(ctx, pool)
		// Both inserts succeed here — the UNIQUE constraint is deferred and
		// only checked at COMMIT.
		if _, err := qr.Exec(ctx, fmt.Sprintf("INSERT INTO %s (val) VALUES ($1)", table), "dup"); err != nil {
			return err
		}
		_, err := qr.Exec(ctx, fmt.Sprintf("INSERT INTO %s (val) VALUES ($1)", table), "dup")
		return err
	})

	So(err, ShouldNotBeNil)
	So(err.Error(), ShouldContainSubstring, "transactor.Commit")

	var pgErr *pgconn.PgError
	So(errors.As(err, &pgErr), ShouldBeTrue)
	So(pgErr.Code, ShouldEqual, "23505") // unique_violation

	var count int
	scanErr := pool.QueryRow(context.Background(), fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
	So(scanErr, ShouldBeNil)
	So(count, ShouldEqual, 0)
}

// rowExists queries pool directly (never through the transaction under test)
func rowExists(t *testing.T, pool *pgxpool.Pool, table, id string) bool {
	t.Helper()

	var found string
	err := pool.QueryRow(context.Background(), fmt.Sprintf("SELECT id FROM %s WHERE id = $1", table), id).Scan(&found)
	if errors.Is(err, pgx.ErrNoRows) {
		return false
	}
	if err != nil {
		t.Fatalf("rowExists: %v", err)
	}
	return true
}

// assertCommitsOnSuccess verifies fn's writes survive after InTransaction returns nil.
func assertCommitsOnSuccess(t *testing.T, pool *pgxpool.Pool, transactor *db.PgxTransactor) {
	t.Helper()
	table := newScratchTable(t, pool)
	const id = "committed-row"

	err := transactor.InTransaction(context.Background(), func(ctx context.Context) error {
		qr := db.FallbackQueryRunner(ctx, pool)
		_, err := qr.Exec(ctx, fmt.Sprintf("INSERT INTO %s (id) VALUES ($1)", table), id)
		return err
	})

	So(err, ShouldBeNil)
	So(rowExists(t, pool, table, id), ShouldBeTrue)
}

// assertRollsBackOnError verifies fn's writes are discarded when fn returns an error.
func assertRollsBackOnError(t *testing.T, pool *pgxpool.Pool, transactor *db.PgxTransactor) {
	t.Helper()
	table := newScratchTable(t, pool)
	const id = "rolled-back-row"
	wantErr := errors.New("fn failed")

	insertThenFail := func(ctx context.Context) error {
		qr := db.FallbackQueryRunner(ctx, pool)
		if _, err := qr.Exec(ctx, fmt.Sprintf("INSERT INTO %s (id) VALUES ($1)", table), id); err != nil {
			return err
		}
		return wantErr
	}
	err := transactor.InTransaction(context.Background(), insertThenFail)

	So(errors.Is(err, wantErr), ShouldBeTrue)
	So(rowExists(t, pool, table, id), ShouldBeFalse)
}

// assertFnObservesTx verifies fn's ctx carries the real transaction (via
// ExtractTx/FallbackQueryRunner), not just the raw pool.
func assertFnObservesTx(t *testing.T, pool *pgxpool.Pool, transactor *db.PgxTransactor) {
	t.Helper()
	table := newScratchTable(t, pool)
	const id = "extract-tx-row"
	var sawTx bool

	err := transactor.InTransaction(context.Background(), func(ctx context.Context) error {
		_, sawTx = db.ExtractTx(ctx)
		qr := db.FallbackQueryRunner(ctx, pool)
		_, err := qr.Exec(ctx, fmt.Sprintf("INSERT INTO %s (id) VALUES ($1)", table), id)
		return err
	})

	So(err, ShouldBeNil)
	So(sawTx, ShouldBeTrue)
	So(rowExists(t, pool, table, id), ShouldBeTrue)
}

// assertBeginTxFailureWrapped verifies InTransaction wraps BeginTx's error and
// never invokes fn when the pool is already closed.
func assertBeginTxFailureWrapped(t *testing.T) {
	t.Helper()
	closedPool := testutil.NewTestPool(t)
	closedPool.Close()
	closedTransactor := db.NewTransactor(closedPool)

	fnCalled := false
	err := closedTransactor.InTransaction(context.Background(), func(_ context.Context) error {
		fnCalled = true
		return nil
	})

	So(fnCalled, ShouldBeFalse)
	So(err, ShouldNotBeNil)
	So(err.Error(), ShouldContainSubstring, "transactor.BeginTx")
}

func TestInTransaction_Integration(t *testing.T) {
	pool := testutil.NewTestPool(t)
	transactor := db.NewTransactor(pool)

	Convey("InTransaction", t, func() {
		Convey("When fn succeeds, the transaction is committed", func() {
			assertCommitsOnSuccess(t, pool, transactor)
		})

		Convey("When fn returns an error, the transaction is rolled back", func() {
			assertRollsBackOnError(t, pool, transactor)
		})

		Convey("fn observes the transaction via ExtractTx/FallbackQueryRunner, not the raw pool", func() {
			assertFnObservesTx(t, pool, transactor)
		})

		Convey("When the pool is already closed, BeginTx fails and the error is wrapped", func() {
			assertBeginTxFailureWrapped(t)
		})

		Convey("When a deferred constraint violation surfaces only at COMMIT time, the error is wrapped and nothing is persisted", func() {
			assertCommitFailureWrapped(t, pool, transactor)
		})
	})
}
