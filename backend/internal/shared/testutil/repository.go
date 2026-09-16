package testutil

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

// MockQueryRunner implements db.QueryRunner via function fields; an unset field panics
// with a descriptive message, surfacing missing test setup immediately.
type MockQueryRunner struct {
	QueryRowFn func(ctx context.Context, sql string, args ...any) pgx.Row
	QueryFn    func(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	ExecFn     func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// AlwaysFailsQuery satisfies MockQueryRunner.QueryFn, failing unconditionally with ErrDBUnexpected.
func AlwaysFailsQuery(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
	return nil, ErrDBUnexpected
}

// AlwaysFailsExec satisfies MockQueryRunner.ExecFn, failing unconditionally with ErrDBUnexpected.
func AlwaysFailsExec(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, ErrDBUnexpected
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

// castAs safely type-asserts a scan destination to *T, panicking with context on failure.
// It backs all the Cast*/CastEnum helpers below — they exist as named, non-generic entry
// points so call sites read naturally (CastStr(v, idx)) without spelling out a type param.
func castAs[T any](v any, idx int) *T {
	t, ok := v.(*T)
	if !ok {
		panic(fmt.Sprintf("dest[%d]: expected *%T, got %T", idx, t, v))
	}
	return t
}

// CastStr safely type-asserts a scan destination to *string, panicking with context on failure.
func CastStr(v any, idx int) *string { return castAs[string](v, idx) }

// CastPtrStr safely type-asserts a scan destination to **string.
func CastPtrStr(v any, idx int) **string { return castAs[*string](v, idx) }

// CastTime safely type-asserts a scan destination to *time.Time.
func CastTime(v any, idx int) *time.Time { return castAs[time.Time](v, idx) }

// CastInt safely type-asserts a scan destination to *int.
func CastInt(v any, idx int) *int { return castAs[int](v, idx) }

// CastFloat64 safely type-asserts a scan destination to *float64.
func CastFloat64(v any, idx int) *float64 { return castAs[float64](v, idx) }

// CastPgtypeDate safely type-asserts a scan destination to *pgtype.Date — used for
// nullable `date` columns since pgx v5 can't scan `date` into *string.
func CastPgtypeDate(v any, idx int) *pgtype.Date { return castAs[pgtype.Date](v, idx) }

// CastBool safely type-asserts a scan destination to *bool.
func CastBool(v any, idx int) *bool { return castAs[bool](v, idx) }

// CastPtrInt safely type-asserts a scan destination to **int, for nullable integer columns.
func CastPtrInt(v any, idx int) **int { return castAs[*int](v, idx) }

// CastPtrTime safely type-asserts a scan destination to **time.Time, for nullable
// timestamptz columns.
func CastPtrTime(v any, idx int) **time.Time { return castAs[*time.Time](v, idx) }

// CastEnum safely type-asserts a scan destination to *T, for domain-specific string-enum
// columns (e.g. CourseStatus, ContentType, UserRole) that can't live in a stdlib Cast* helper.
func CastEnum[T any](v any, idx int) *T { return castAs[T](v, idx) }

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
func (r *MockRows) Close() {
	// intentionally empty — see doc comment above
}

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
