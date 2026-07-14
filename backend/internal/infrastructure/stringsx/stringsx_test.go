package stringsx_test

import (
	"strings"
	"testing"

	"learnflow_backend/internal/infrastructure/stringsx"

	. "github.com/smartystreets/goconvey/convey"
)

func TestToPascalFromSeparated(t *testing.T) {
	Convey("ToPascalFromSeparated", t, func() {
		Convey("converts snake_case", func() {
			So(stringsx.ToPascalFromSeparated("access_token"), ShouldEqual, "AccessToken")
		})
		Convey("converts kebab-case", func() {
			So(stringsx.ToPascalFromSeparated("x-api-key"), ShouldEqual, "XApiKey")
		})
		Convey("converts mixed snake and kebab", func() {
			So(stringsx.ToPascalFromSeparated("proxy-authorization_basic"), ShouldEqual, "ProxyAuthorizationBasic")
		})
		Convey("single word is title-cased", func() {
			So(stringsx.ToPascalFromSeparated("token"), ShouldEqual, "Token")
		})
		Convey("empty string stays empty", func() {
			So(stringsx.ToPascalFromSeparated(""), ShouldEqual, "")
		})
		Convey("uppercased input is lowercased except first rune", func() {
			So(stringsx.ToPascalFromSeparated("ACCESS_TOKEN"), ShouldEqual, "AccessToken")
		})
	})
}

func TestTruncateString(t *testing.T) {
	Convey("TruncateString", t, func() {
		Convey("When limit is zero or negative", func() {
			Convey("returns empty string", func() {
				So(stringsx.TruncateString("hello", 0), ShouldEqual, "")
				So(stringsx.TruncateString("hello", -1), ShouldEqual, "")
			})
		})

		Convey("When string fits within limit", func() {
			Convey("returns the original string unchanged", func() {
				So(stringsx.TruncateString("hello", 10), ShouldEqual, "hello")
				So(stringsx.TruncateString("hello", 5), ShouldEqual, "hello")
			})
		})

		Convey("When string exceeds limit", func() {
			Convey("truncates and appends marker", func() {
				result := stringsx.TruncateString("hello world", 5)
				So(result, ShouldEqual, "hello...[TRUNCATED]")
			})
		})

		Convey("When string contains multi-byte Cyrillic characters", func() {
			s := strings.Repeat("я", 10)
			Convey("truncates by rune count not bytes", func() {
				result := stringsx.TruncateString(s, 5)
				So(result, ShouldEqual, "яяяяя...[TRUNCATED]")
			})
		})

		Convey("When limit equals exact rune count", func() {
			Convey("returns original without marker", func() {
				So(stringsx.TruncateString("abcde", 5), ShouldEqual, "abcde")
			})
		})
	})
}
