//go:build integration

package migrations

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"

	"learnflow_backend/internal/infrastructure/db"

	"github.com/golang-migrate/migrate/v4"
	// Registers the pgx/v5 database driver with golang-migrate's driver registry.
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"

	. "github.com/smartystreets/goconvey/convey"
)

// newThrowawayDatabase creates a uniquely-named database on the same Postgres instance
// as the shared integration test DB, and returns a DSN pointing at it plus a cleanup
// func that drops it. Migrations up/down cycles run destructively (they create and drop
// every table in the schema); running them against the shared learnflow_test database
// would corrupt state for every other integration test in the suite.
func newThrowawayDatabase(t *testing.T) (dsn string, cleanup func()) {
	t.Helper()

	adminDSN, err := db.BuildDSNFromEnv()
	if err != nil {
		t.Fatalf("newThrowawayDatabase: %v", err)
	}

	adminPool, err := pgxpool.New(context.Background(), adminDSN)
	if err != nil {
		t.Fatalf("newThrowawayDatabase: connect: %v", err)
	}
	defer adminPool.Close()

	buf := make([]byte, 4)
	if _, err := rand.Read(buf); err != nil {
		t.Fatalf("newThrowawayDatabase: %v", err)
	}
	// dbName is derived only from hex-encoded random bytes (never external input), so
	// string formatting here carries no SQL-injection risk — CREATE/DROP DATABASE don't
	// support parameterized identifiers at all, this is the only way to issue them.
	dbName := "migrate_rollback_test_" + hex.EncodeToString(buf)

	ctx := context.Background()
	if _, err := adminPool.Exec(ctx, fmt.Sprintf(`CREATE DATABASE %s`, dbName)); err != nil {
		t.Fatalf("newThrowawayDatabase: create database: %v", err)
	}

	u, err := url.Parse(adminDSN)
	if err != nil {
		t.Fatalf("newThrowawayDatabase: parse dsn: %v", err)
	}
	u.Path = "/" + dbName
	dsn = u.String()

	cleanup = func() {
		dropPool, err := pgxpool.New(context.Background(), adminDSN)
		if err != nil {
			t.Logf("newThrowawayDatabase cleanup: connect: %v", err)
			return
		}
		defer dropPool.Close()
		// WITH (FORCE) disconnects any lingering sessions (e.g. migrate's own pool, if
		// its Close() hasn't fully released connections yet) so DROP DATABASE can't hang.
		if _, err := dropPool.Exec(context.Background(), fmt.Sprintf(`DROP DATABASE IF EXISTS %s WITH (FORCE)`, dbName)); err != nil {
			t.Logf("newThrowawayDatabase cleanup: drop database: %v", err)
		}
	}

	return dsn, cleanup
}

// stepDownToBottom steps m down one migration at a time until it hits the bottom
// (ErrNilVersion / ErrNotExist), failing the test if that never happens within 100
// steps. Returns the number of steps taken.
func stepDownToBottom(t *testing.T, m *migrate.Migrate) int {
	t.Helper()

	steps := 0
	for {
		err := m.Steps(-1)
		if errors.Is(err, migrate.ErrNilVersion) || errors.Is(err, os.ErrNotExist) {
			return steps
		}
		So(err, ShouldBeNil)
		steps++
		if steps > 100 {
			t.Fatal("stepDownToBottom: down loop did not terminate")
		}
	}
}

func newMigrator(t *testing.T, dsn string) *migrate.Migrate {
	t.Helper()

	d, err := iofs.New(FS, ".")
	if err != nil {
		t.Fatalf("newMigrator: iofs source: %v", err)
	}

	migrateDSN := strings.Replace(dsn, "postgres://", "pgx5://", 1)
	m, err := migrate.NewWithSourceInstance("iofs", d, migrateDSN)
	if err != nil {
		t.Fatalf("newMigrator: %v", err)
	}
	return m
}

func TestMigrationsUpDownUp_Integration(t *testing.T) {
	Convey("Given a fresh, empty throwaway database", t, func() {
		dsn, cleanup := newThrowawayDatabase(t)
		defer cleanup()

		m := newMigrator(t, dsn)
		defer func() { _, _ = m.Close() }()

		Convey("When applying all migrations up", func() {
			err := m.Up()
			So(err, ShouldBeNil)

			version, dirty, err := m.Version()
			So(err, ShouldBeNil)
			So(dirty, ShouldBeFalse)
			So(version, ShouldBeGreaterThan, uint(0))

			Convey("When rolling all migrations back down", func() {
				err := m.Down()
				So(err, ShouldBeNil)

				_, dirty, err := m.Version()
				So(errors.Is(err, migrate.ErrNilVersion), ShouldBeTrue)
				So(dirty, ShouldBeFalse)

				Convey("When re-applying all migrations up again", func() {
					err := m.Up()
					So(err, ShouldBeNil)

					version, dirty, err := m.Version()
					So(err, ShouldBeNil)
					So(dirty, ShouldBeFalse)
					So(version, ShouldBeGreaterThan, uint(0))
				})
			})
		})
	})
}

func TestMigrationsStepDownUpOneAtATime_Integration(t *testing.T) {
	Convey("Given a database with every migration applied", t, func() {
		dsn, cleanup := newThrowawayDatabase(t)
		defer cleanup()

		m := newMigrator(t, dsn)
		defer func() { _, _ = m.Close() }()

		So(m.Up(), ShouldBeNil)
		topVersion, _, err := m.Version()
		So(err, ShouldBeNil)

		Convey("When stepping down one migration at a time to the bottom", func() {
			steps := stepDownToBottom(t, m)
			So(steps, ShouldBeGreaterThan, 0)

			Convey("When stepping back up one migration at a time to the top", func() {
				for i := 0; i < steps; i++ {
					So(m.Steps(1), ShouldBeNil)
				}

				version, dirty, err := m.Version()
				So(err, ShouldBeNil)
				So(dirty, ShouldBeFalse)
				So(version, ShouldEqual, topVersion)
			})
		})
	})
}
