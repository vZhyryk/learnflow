package testutil

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

// MockQueryRunner implements db.QueryRunner (internal/infrastructure/db) via
// function fields. Any unset field panics with a descriptive message on use,
// which surfaces missing test setup immediately instead of nil-pointer-panicking.
type MockQueryRunner struct {
	QueryRowFn func(ctx context.Context, sql string, args ...any) pgx.Row
	QueryFn    func(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	ExecFn     func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// QueryRow delegates to QueryRowFn, panicking if it was not set for this test.
func (m *MockQueryRunner) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if m.QueryRowFn == nil {
		panic("testutil.MockQueryRunner.QueryRowFn not set")
	}
	return m.QueryRowFn(ctx, sql, args...)
}

// Query delegates to QueryFn, panicking if it was not set for this test.
func (m *MockQueryRunner) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if m.QueryFn == nil {
		panic("testutil.MockQueryRunner.QueryFn not set")
	}
	return m.QueryFn(ctx, sql, args...)
}

// Exec delegates to ExecFn, panicking if it was not set for this test.
func (m *MockQueryRunner) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if m.ExecFn == nil {
		panic("testutil.MockQueryRunner.ExecFn not set")
	}
	return m.ExecFn(ctx, sql, args...)
}

// MockRow implements pgx.Row for controlled Scan injection.
type MockRow struct {
	ScanFn func(dest ...any) error
}

// Scan delegates to ScanFn.
func (r *MockRow) Scan(dest ...any) error { return r.ScanFn(dest...) }

// CastStr safely type-asserts a scan destination to *string, panicking with context on failure.
func CastStr(v any, idx int) *string {
	s, ok := v.(*string)
	if !ok {
		panic(fmt.Sprintf("dest[%d]: expected *string, got %T", idx, v))
	}
	return s
}

// CastPtrStr safely type-asserts a scan destination to **string.
func CastPtrStr(v any, idx int) **string {
	s, ok := v.(**string)
	if !ok {
		panic(fmt.Sprintf("dest[%d]: expected **string, got %T", idx, v))
	}
	return s
}

// CastTime safely type-asserts a scan destination to *time.Time.
func CastTime(v any, idx int) *time.Time {
	s, ok := v.(*time.Time)
	if !ok {
		panic(fmt.Sprintf("dest[%d]: expected *time.Time, got %T", idx, v))
	}
	return s
}

// CastInt safely type-asserts a scan destination to *int.
func CastInt(v any, idx int) *int {
	s, ok := v.(*int)
	if !ok {
		panic(fmt.Sprintf("dest[%d]: expected *int, got %T", idx, v))
	}
	return s
}

// CastPgtypeDate safely type-asserts a scan destination to *pgtype.Date —
// the scan target repositories use for nullable `date` columns (pgx v5 has
// no native scan support for `date` into *string/**string).
func CastPgtypeDate(v any, idx int) *pgtype.Date {
	d, ok := v.(*pgtype.Date)
	if !ok {
		panic(fmt.Sprintf("dest[%d]: expected *pgtype.Date, got %T", idx, v))
	}
	return d
}

// MockRows implements pgx.Rows for controlled multi-row Scan injection in
// repository/worker tests. Rows are consumed front-to-back by successive Scan calls.
type MockRows struct {
	Rows    []*MockRow
	RowsErr error // returned by Err() after iteration completes
}

// Next reports whether there are more rows to scan.
func (r *MockRows) Next() bool {
	return len(r.Rows) > 0
}

// Scan delegates to the next MockRow's ScanFn and advances the cursor.
func (r *MockRows) Scan(dest ...any) error {
	err := r.Rows[0].Scan(dest...)
	r.Rows = r.Rows[1:]
	return err
}

// Close is a no-op; MockRows has no underlying connection to release.
func (r *MockRows) Close() {}

// Err returns RowsErr, the error to surface after iteration completes.
func (r *MockRows) Err() error { return r.RowsErr }

// CommandTag returns an empty tag; not meaningful for a mock.
func (r *MockRows) CommandTag() pgconn.CommandTag { return pgconn.CommandTag{} }

// FieldDescriptions returns nil; not meaningful for a mock.
func (r *MockRows) FieldDescriptions() []pgconn.FieldDescription { return nil }

// Values returns nil; not meaningful for a mock.
func (r *MockRows) Values() ([]any, error) { return nil, nil }

// RawValues returns nil; not meaningful for a mock.
func (r *MockRows) RawValues() [][]byte { return nil }

// Conn returns nil; not meaningful for a mock.
func (r *MockRows) Conn() *pgx.Conn { return nil }

// NoopTransactor runs fn(ctx) immediately, without a real database transaction.
// Satisfies any InTransaction(ctx context.Context, fn func(context.Context) error) error interface.
type NoopTransactor struct{}

// InTransaction runs fn(ctx) immediately and returns its result.
func (NoopTransactor) InTransaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}
