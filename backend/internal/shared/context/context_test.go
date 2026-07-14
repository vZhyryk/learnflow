package appcontext_test

import (
	"context"
	"strings"
	"testing"
	"time"

	authdomain "learnflow_backend/internal/auth/domain"
	appcontext "learnflow_backend/internal/shared/context"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewRequestID(t *testing.T) {
	Convey("NewRequestID", t, func() {
		id := appcontext.NewRequestID()

		Convey("has UUID v4 format (8-4-4-4-12)", func() {
			parts := strings.Split(id, "-")
			So(len(parts), ShouldEqual, 5)
			So(len(parts[0]), ShouldEqual, 8)
			So(len(parts[1]), ShouldEqual, 4)
			So(len(parts[2]), ShouldEqual, 4)
			So(len(parts[3]), ShouldEqual, 4)
			So(len(parts[4]), ShouldEqual, 12)
		})

		Convey("version nibble is 4", func() {
			parts := strings.Split(id, "-")
			So(parts[2][0], ShouldEqual, '4')
		})

		Convey("each call produces unique value", func() {
			So(id, ShouldNotEqual, appcontext.NewRequestID())
		})
	})
}

func TestRequestIDContext(t *testing.T) {
	Convey("RequestID context", t, func() {
		Convey("roundtrip: set and get returns same value", func() {
			ctx := appcontext.WithRequestID(context.Background(), "req-abc")
			So(appcontext.RequestIDFromContext(ctx), ShouldEqual, "req-abc")
		})

		Convey("missing key returns empty string", func() {
			So(appcontext.RequestIDFromContext(context.Background()), ShouldEqual, "")
		})
	})
}

func TestIPAddressContext(t *testing.T) {
	Convey("IPAddress context", t, func() {
		Convey("roundtrip: set and get returns same value", func() {
			ctx := appcontext.WithIPAddress(context.Background(), "1.2.3.4")
			So(appcontext.IPAddressFromContext(ctx), ShouldEqual, "1.2.3.4")
		})

		Convey("missing key returns empty string", func() {
			So(appcontext.IPAddressFromContext(context.Background()), ShouldEqual, "")
		})
	})
}

func TestUserContext(t *testing.T) {
	Convey("User context", t, func() {
		user := &authdomain.User{ID: "user-123", Role: authdomain.RoleUser}

		Convey("roundtrip: set and get returns same user", func() {
			ctx := appcontext.WithUser(context.Background(), user)
			got, ok := appcontext.UserFromContext(ctx)
			So(ok, ShouldBeTrue)
			So(got.ID, ShouldEqual, "user-123")
			So(got.Role, ShouldEqual, authdomain.RoleUser)
		})

		Convey("missing key returns nil and false", func() {
			got, ok := appcontext.UserFromContext(context.Background())
			So(ok, ShouldBeFalse)
			So(got, ShouldBeNil)
		})

		Convey("nil user stored returns nil and false", func() {
			ctx := appcontext.WithUser(context.Background(), nil)
			got, ok := appcontext.UserFromContext(ctx)
			So(ok, ShouldBeFalse)
			So(got, ShouldBeNil)
		})
	})
}

func TestJTIContext(t *testing.T) {
	Convey("JTI context", t, func() {
		Convey("roundtrip: set and get returns same value", func() {
			ctx := appcontext.WithJTI(context.Background(), "jti-xyz")
			So(appcontext.JTIFromContext(ctx), ShouldEqual, "jti-xyz")
		})

		Convey("missing key returns empty string", func() {
			So(appcontext.JTIFromContext(context.Background()), ShouldEqual, "")
		})
	})
}

func TestAccessTokenExpiresAtContext(t *testing.T) {
	Convey("AccessTokenExpiresAt context", t, func() {
		now := time.Now().Truncate(time.Second)

		Convey("roundtrip: set and get returns same time", func() {
			ctx := appcontext.WithAccessTokenExpiresAt(context.Background(), now)
			So(appcontext.AccessTokenExpiresAtFromContext(ctx), ShouldEqual, now)
		})

		Convey("missing key returns zero time", func() {
			So(appcontext.AccessTokenExpiresAtFromContext(context.Background()), ShouldEqual, time.Time{})
		})
	})
}
