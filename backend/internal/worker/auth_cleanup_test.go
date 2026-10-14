package worker

import (
	"context"
	"learnflow_backend/internal/shared/testutil"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	. "github.com/smartystreets/goconvey/convey"
)

func TestTokenCleanupWorkers(t *testing.T) {
	Convey("NewTokenCleanupWorkers", t, func() {
		var gotQueries []string
		runner := &testutil.MockQueryRunner{
			ExecFn: func(_ context.Context, sql string, _ ...any) (pgconn.CommandTag, error) {
				gotQueries = append(gotQueries, sql)
				return pgconn.NewCommandTag("DELETE 0"), nil
			},
		}

		workers := NewTokenCleanupWorkers(runner, testutil.NewTestLogger(), time.Hour)

		Convey("it returns one worker per token table, each deleting only expired rows", func() {
			So(workers, ShouldHaveLength, len(tokenTables))

			for _, w := range workers {
				cw, ok := w.(*CleanupWorker[TokenCleanupWorker])
				So(ok, ShouldBeTrue)
				cw.poll(context.Background())
			}

			So(gotQueries, ShouldHaveLength, len(tokenTables))
			for i, table := range tokenTables {
				So(gotQueries[i], ShouldContainSubstring, "DELETE FROM "+table)
				So(gotQueries[i], ShouldContainSubstring, "expires_at < now()")
				So(gotQueries[i], ShouldContainSubstring, "LIMIT 1000")
			}
		})
	})
}

func TestSessionCleanupWorker(t *testing.T) {
	Convey("NewSessionCleanupWorker", t, func() {
		var gotQuery string
		runner := &testutil.MockQueryRunner{
			ExecFn: func(_ context.Context, sql string, _ ...any) (pgconn.CommandTag, error) {
				gotQuery = sql
				return pgconn.NewCommandTag("DELETE 0"), nil
			},
		}

		NewSessionCleanupWorker(runner, testutil.NewTestLogger(), time.Hour).poll(context.Background())

		So(gotQuery, ShouldContainSubstring, "DELETE FROM user_sessions")
		So(gotQuery, ShouldContainSubstring, "interval '"+sessionRetention+"'")
	})
}
