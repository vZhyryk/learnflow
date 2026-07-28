package repository

import (
	"context"
	"testing"

	"learnflow_backend/internal/shared/testutil"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewBaseRepository(t *testing.T) {
	Convey("NewBaseRepository", t, func() {
		Convey("returns a non-nil BaseRepository", func() {
			rep := NewBaseRepository(nil)
			So(rep, ShouldNotBeNil)
		})
	})
}

func TestBaseRepositoryQueryRunner(t *testing.T) {
	Convey("BaseRepository.QueryRunner", t, func() {
		runner := &testutil.MockQueryRunner{}
		rep := &BaseRepository{DB: runner}

		Convey("no active transaction in ctx falls back to DB", func() {
			So(rep.QueryRunner(context.Background()), ShouldEqual, runner)
		})
	})
}
