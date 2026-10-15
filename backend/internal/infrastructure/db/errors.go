package db

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// IsUniqueViolation reports whether err is a Postgres unique_violation (23505) on the
// given constraint — the DB-level backstop for check-then-insert/update races.
func IsUniqueViolation(err error, constraintName string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == constraintName
}

// IsCheckViolation reports whether err is a Postgres check_violation (23514) on the
// given constraint — used to map a CHECK-constraint failure (e.g. a polymorphic
// entity_type/entity_id pairing rule) to a domain validation error.
func IsCheckViolation(err error, constraintName string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23514" && pgErr.ConstraintName == constraintName
}
