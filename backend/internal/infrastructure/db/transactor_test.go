package db

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"

	. "github.com/smartystreets/goconvey/convey"
)

// fakeTx is a minimal pgx.Tx stand-in. Embedding the interface as a nil value
// satisfies pgx.Tx without implementing every method — ExtractTx only type-asserts,
// it never calls through the interface, so unimplemented methods are never reached.
type fakeTx struct {
	pgx.Tx
}

func TestExtractTx(t *testing.T) {
	Convey("Given a context", t, func() {
		Convey("When no transaction is stored", func() {
			tx, ok := ExtractTx(context.Background())
			So(ok, ShouldBeFalse)
			So(tx, ShouldBeNil)
		})

		Convey("When the context holds an unrelated value under a different key", func() {
			ctx := context.WithValue(context.Background(), struct{ unrelated string }{"key"}, "value")
			tx, ok := ExtractTx(ctx)
			So(ok, ShouldBeFalse)
			So(tx, ShouldBeNil)
		})

		Convey("When a transaction is stored via InTransaction's context key", func() {
			want := &fakeTx{}
			ctx := context.WithValue(context.Background(), txKey{}, pgx.Tx(want))
			got, ok := ExtractTx(ctx)
			So(ok, ShouldBeTrue)
			So(got, ShouldEqual, want)
		})
	})
}

// fakeQueryRunner is a minimal QueryRunner stand-in used only to verify identity
// (which instance FallbackQueryRunner returns), never called through. It can't
// reuse testutil.MockQueryRunner here: testutil imports this db package
// elsewhere (NewTestPool etc.), so importing testutil back from a db-package
// test file would be an import cycle.
type fakeQueryRunner struct {
	QueryRunner
}

func TestFallbackQueryRunner(t *testing.T) {
	Convey("Given a fallback QueryRunner", t, func() {
		fallback := &fakeQueryRunner{}

		Convey("When no transaction is stored in ctx", func() {
			got := FallbackQueryRunner(context.Background(), fallback)
			So(got, ShouldEqual, fallback)
		})

		Convey("When a transaction is stored in ctx", func() {
			tx := &fakeTx{}
			ctx := context.WithValue(context.Background(), txKey{}, pgx.Tx(tx))
			got := FallbackQueryRunner(ctx, fallback)
			So(got, ShouldEqual, tx)
			So(got, ShouldNotEqual, fallback)
		})
	})
}
