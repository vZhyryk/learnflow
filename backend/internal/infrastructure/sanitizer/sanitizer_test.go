package sanitizer_test

import (
	"testing"

	"learnflow_backend/internal/infrastructure/sanitizer"

	. "github.com/smartystreets/goconvey/convey"
)

func newS() *sanitizer.Sanitizer {
	return sanitizer.NewSanitizer("***", 200, map[string]struct{}{
		"password": {},
		"token":    {},
		"bearer":   {},
	})
}

// --- NewSanitizer defaults ---

func TestNewSanitizerDefaults(t *testing.T) {
	Convey("NewSanitizer", t, func() {
		Convey("empty redactedValue defaults to [REDACTED]", func() {
			s := sanitizer.NewSanitizer("", 100, map[string]struct{}{"password": {}})
			out, ok := s.Sanitize(map[string]any{"password": "secret"}).(map[string]any)
			So(ok, ShouldBeTrue)
			So(out["password"], ShouldEqual, "[REDACTED]")
		})

		Convey("zero maxStringLength defaults to 2000", func() {
			s := sanitizer.NewSanitizer("***", 0, map[string]struct{}{})
			long := make([]rune, 1999)
			for i := range long {
				long[i] = 'a'
			}
			So(s.SanitizeString(string(long)), ShouldHaveLength, 1999)
		})

		Convey("nil sensitiveKeys defaults to DefaultSensitiveKeys", func() {
			s := sanitizer.NewSanitizer("***", 100, nil)
			So(s.IsSensitiveKey("password"), ShouldBeTrue)
			So(s.IsSensitiveKey("authorization"), ShouldBeTrue)
		})
	})
}

// --- IsSensitiveKey ---

func TestIsSensitiveKey(t *testing.T) {
	s := newS()

	Convey("IsSensitiveKey", t, func() {
		Convey("exact match is sensitive", func() {
			So(s.IsSensitiveKey("password"), ShouldBeTrue)
		})
		Convey("case-insensitive match is sensitive", func() {
			So(s.IsSensitiveKey("PASSWORD"), ShouldBeTrue)
			So(s.IsSensitiveKey("Password"), ShouldBeTrue)
		})
		Convey("underscore is stripped during normalization", func() {
			// "pass_word" normalizes to "password" which IS in the list
			So(s.IsSensitiveKey("pass_word"), ShouldBeTrue)
			So(s.IsSensitiveKey("token"), ShouldBeTrue)
		})
		Convey("non-sensitive key returns false", func() {
			So(s.IsSensitiveKey("username"), ShouldBeFalse)
			So(s.IsSensitiveKey("email"), ShouldBeFalse)
		})
	})
}

// --- SanitizeString ---

func TestSanitizeString(t *testing.T) {
	s := newS()

	Convey("SanitizeString", t, func() {
		Convey("plain string without sensitive data passes through", func() {
			So(s.SanitizeString("hello world"), ShouldEqual, "hello world")
		})

		Convey("string with inline password=value is redacted", func() {
			result := s.SanitizeString("user logged in password=secret123 ok")
			So(result, ShouldContainSubstring, "***")
			So(result, ShouldNotContainSubstring, "secret123")
		})

		Convey("string exceeding maxLen is truncated", func() {
			long := make([]rune, 201)
			for i := range long {
				long[i] = 'x'
			}
			result := s.SanitizeString(string(long))
			So(result, ShouldContainSubstring, "[TRUNCATED]")
		})
	})
}

// --- SanitizeURL ---

func TestSanitizeURL(t *testing.T) {
	s := newS()

	Convey("SanitizeURL", t, func() {
		Convey("URL with sensitive query param: original value is removed", func() {
			raw := "https://api.example.com/v1?token=abc123&page=2"
			result := s.SanitizeURL(raw)
			So(result, ShouldNotContainSubstring, "abc123")
			So(result, ShouldContainSubstring, "page=2")
			So(result, ShouldContainSubstring, "token=")
		})

		Convey("URL without sensitive params is unchanged (except possible reordering)", func() {
			raw := "https://api.example.com/v1?page=1&sort=asc"
			result := s.SanitizeURL(raw)
			So(result, ShouldContainSubstring, "page=1")
			So(result, ShouldContainSubstring, "sort=asc")
		})

		Convey("string with invalid URL escape falls back to string sanitization", func() {
			// "%zz" is an invalid percent-encoded sequence that url.Parse rejects
			result := s.SanitizeURL("https://host/%zz?token=leak")
			So(result, ShouldNotContainSubstring, "leak")
		})
	})
}

// --- Sanitize (type dispatch) ---

func TestSanitize(t *testing.T) {
	s := newS()

	Convey("Sanitize", t, func() {
		Convey("nil returns nil", func() {
			So(s.Sanitize(nil), ShouldBeNil)
		})

		Convey("string is sanitized", func() {
			So(s.Sanitize("hello"), ShouldEqual, "hello")
		})

		Convey("map[string]any: sensitive key value is redacted", func() {
			in := map[string]any{"password": "s3cr3t", "name": "Alice"}
			out, ok := s.Sanitize(in).(map[string]any)
			So(ok, ShouldBeTrue)
			So(out["password"], ShouldEqual, "***")
			So(out["name"], ShouldEqual, "Alice")
		})

		Convey("map[string]string: sensitive key value is redacted", func() {
			in := map[string]string{"token": "tok123", "email": "a@b.com"}
			out, ok := s.Sanitize(in).(map[string]any)
			So(ok, ShouldBeTrue)
			So(out["token"], ShouldEqual, "***")
			So(out["email"], ShouldEqual, "a@b.com")
		})

		Convey("[]any: each element is sanitized", func() {
			in := []any{"hello", "password=s3cr3t"}
			out, ok := s.Sanitize(in).([]any)
			So(ok, ShouldBeTrue)
			So(out[0], ShouldEqual, "hello")
			elem, ok2 := out[1].(string)
			So(ok2, ShouldBeTrue)
			So(elem, ShouldContainSubstring, "***")
		})

		Convey("[]string: each element is sanitized", func() {
			in := []string{"ok", "token=leak"}
			out, ok := s.Sanitize(in).([]any)
			So(ok, ShouldBeTrue)
			So(out[0], ShouldEqual, "ok")
			elem, ok2 := out[1].(string)
			So(ok2, ShouldBeTrue)
			So(elem, ShouldContainSubstring, "***")
		})

		Convey("struct: fields are reflected and sensitive ones redacted", func() {
			type creds struct {
				Password string
				Username string
			}
			in := creds{Password: "s3cr3t", Username: "alice"}
			out, ok := s.Sanitize(in).(map[string]any)
			So(ok, ShouldBeTrue)
			So(out["Password"], ShouldEqual, "***")
			So(out["Username"], ShouldEqual, "alice")
		})

		Convey("pointer to string: dereferenced and sanitized", func() {
			val := "hello"
			So(s.Sanitize(&val), ShouldEqual, "hello")
		})

		Convey("nil pointer returns nil", func() {
			var p *string
			So(s.Sanitize(p), ShouldBeNil)
		})

		Convey("integer passes through unchanged", func() {
			So(s.Sanitize(42), ShouldEqual, 42)
		})
	})
}

// --- SanitizeMap ---

func TestSanitizeMap(t *testing.T) {
	s := newS()

	Convey("SanitizeMap", t, func() {
		Convey("nil input returns nil", func() {
			So(sanitizer.SanitizeMap(s, map[string]any(nil)), ShouldBeNil)
		})

		Convey("sensitive key is redacted, others preserved", func() {
			in := map[string]any{"token": "abc", "user": "bob"}
			out := sanitizer.SanitizeMap(s, in)
			So(out["token"], ShouldEqual, "***")
			So(out["user"], ShouldEqual, "bob")
		})
	})
}

// --- SanitizeSlice ---

func TestSanitizeSlice(t *testing.T) {
	s := newS()

	Convey("SanitizeSlice", t, func() {
		Convey("nil input returns nil", func() {
			So(sanitizer.SanitizeSlice(s, []any(nil)), ShouldBeNil)
		})

		Convey("each element is sanitized", func() {
			in := []string{"ok", "token=leak"}
			out := sanitizer.SanitizeSlice(s, in)
			So(out[0], ShouldEqual, "ok")
			elem, ok := out[1].(string)
			So(ok, ShouldBeTrue)
			So(elem, ShouldContainSubstring, "***")
		})
	})
}

// --- KeyVariants ---

func TestKeyVariants(t *testing.T) {
	s := newS()

	Convey("KeyVariants", t, func() {
		Convey("plain word gets lowercase, uppercase, title-case", func() {
			variants := s.KeyVariants("token")
			So(variants, ShouldContain, "token")
			So(variants, ShouldContain, "TOKEN")
			So(variants, ShouldContain, "Token")
		})

		Convey("snake_case word also gets PascalCase", func() {
			variants := s.KeyVariants("access_token")
			So(variants, ShouldContain, "access_token")
			So(variants, ShouldContain, "ACCESS_TOKEN")
			So(variants, ShouldContain, "AccessToken")
		})

		Convey("no duplicates", func() {
			variants := s.KeyVariants("TOKEN")
			seen := make(map[string]int)
			for _, v := range variants {
				seen[v]++
			}
			for _, count := range seen {
				So(count, ShouldEqual, 1)
			}
		})
	})
}

// --- MaskAllWithMarker ---

func TestMaskAllWithMarker(t *testing.T) {
	s := newS()

	Convey("MaskAllWithMarker", t, func() {
		Convey("When marker is not present", func() {
			Convey("returns original string", func() {
				So(s.MaskAllWithMarker("no secret here", "password="), ShouldEqual, "no secret here")
			})
		})

		Convey("When marker is present once", func() {
			Convey("replaces value after marker", func() {
				result := s.MaskAllWithMarker("password=abc123 rest", "password=")
				So(result, ShouldContainSubstring, "password=***")
				So(result, ShouldNotContainSubstring, "abc123")
				So(result, ShouldContainSubstring, " rest")
			})
		})

		Convey("When value reaches end of string", func() {
			Convey("replaces value at end", func() {
				result := s.MaskAllWithMarker("password=secret", "password=")
				So(result, ShouldEqual, "password=***")
			})
		})
	})
}

// --- SanitizePath ---

func TestSanitizePath(t *testing.T) {
	s := newS()

	Convey("SanitizePath", t, func() {
		Convey("Opaque token segment is redacted", func() {
			result := s.SanitizePath("/auth/reset-password/eyJhbGciOiJIUzI1NiJ9.super-secret-token")
			So(result, ShouldEqual, "/auth/reset-password/***")
		})

		Convey("UUID resource ID segment is left untouched", func() {
			path := "/users/550e8400-e29b-41d4-a716-446655440000/profile"
			So(s.SanitizePath(path), ShouldEqual, path)
		})

		Convey("Short segments (route names, ordinary slugs) are left untouched", func() {
			path := "/auth/reset-password"
			So(s.SanitizePath(path), ShouldEqual, path)
		})

		Convey("Multiple opaque segments in one path are each redacted", func() {
			result := s.SanitizePath("/a/aaaaaaaaaaaaaaaaaaaaaaaa/b/bbbbbbbbbbbbbbbbbbbbbbbb")
			So(result, ShouldEqual, "/a/***/b/***")
		})

		Convey("Empty path is left untouched", func() {
			So(s.SanitizePath(""), ShouldEqual, "")
		})
	})
}
