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
