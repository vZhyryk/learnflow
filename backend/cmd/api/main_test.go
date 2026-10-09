package main

import (
	"learnflow_backend/cmd/api/app"
	"strings"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	validJWTSecret    = "0123456789abcdef0123456789abcdef"
	previousJWTSecret = "fedcba9876543210fedcba9876543210"
)

func setValidJWTEnv(t *testing.T) {
	t.Helper()
	t.Setenv("JWT_SECRET", validJWTSecret)
	t.Setenv("JWT_SECRET_PREV", "")
	t.Setenv("JWT_ISSUER", "learnflow")
	t.Setenv("JWT_AUDIENCE", "learnflow-users")
}

func TestGetJWTConfig(t *testing.T) {
	Convey("getJWTConfig", t, func() {
		Convey("When all values are valid, it fills the config", func() {
			setValidJWTEnv(t)
			t.Setenv("JWT_SECRET_PREV", previousJWTSecret)
			cfg := &app.Config{}

			So(getJWTConfig(cfg), ShouldBeNil)
			So(cfg.Secret.JWTSecret, ShouldEqual, validJWTSecret)
			So(cfg.Secret.JWTSecretPrev, ShouldEqual, previousJWTSecret)
			So(cfg.Secret.JWTIssuer, ShouldEqual, "learnflow")
			So(cfg.Secret.JWTAudience, ShouldEqual, "learnflow-users")
		})

		cases := map[string]struct {
			key, value, wantErr string
		}{
			"empty secret":      {"JWT_SECRET", "", "JWT_SECRET cannot be empty"},
			"short secret":      {"JWT_SECRET", strings.Repeat("a", 31), "at least 32 bytes"},
			"short prev secret": {"JWT_SECRET_PREV", strings.Repeat("a", 31), "JWT_SECRET_PREV must be at least 32 bytes"},
			"empty issuer":      {"JWT_ISSUER", "", "JWT_ISSUER cannot be empty"},
			"empty audience":    {"JWT_AUDIENCE", "", "JWT_AUDIENCE cannot be empty"},
		}
		for name, tc := range cases {
			Convey("When "+name+", it fails startup", func() {
				setValidJWTEnv(t)
				t.Setenv(tc.key, tc.value)

				err := getJWTConfig(&app.Config{})

				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, tc.wantErr)
			})
		}

		Convey("When the secret is exactly 32 bytes, it is accepted", func() {
			setValidJWTEnv(t)
			t.Setenv("JWT_SECRET", strings.Repeat("a", 32))

			So(getJWTConfig(&app.Config{}), ShouldBeNil)
		})
	})
}

func TestParseTrustedProxies(t *testing.T) {
	Convey("parseTrustedProxies", t, func() {
		Convey("When the input is empty, it returns no proxies", func() {
			got, err := parseTrustedProxies("")

			So(err, ShouldBeNil)
			So(got, ShouldBeEmpty)
		})

		Convey("When the input has several CIDRs with spaces and empty items, it parses each", func() {
			got, err := parseTrustedProxies(" 10.0.0.0/8, ,192.168.1.0/24 ,")

			So(err, ShouldBeNil)
			So(got, ShouldHaveLength, 2)
			So(got[0].String(), ShouldEqual, "10.0.0.0/8")
			So(got[1].String(), ShouldEqual, "192.168.1.0/24")
		})

		Convey("When a CIDR is host-bits-set, it is normalised to the network", func() {
			got, err := parseTrustedProxies("10.1.2.3/8")

			So(err, ShouldBeNil)
			So(got[0].String(), ShouldEqual, "10.0.0.0/8")
		})

		for _, bad := range []string{"not-a-cidr", "10.0.0.1", "10.0.0.0/33", "10.0.0.0/8,oops"} {
			Convey("When the input is invalid ("+bad+"), it returns an error naming the entry", func() {
				got, err := parseTrustedProxies(bad)

				So(got, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "invalid CIDR")
			})
		}
	})
}

func TestGetCorsTrustedOrigins(t *testing.T) {
	Convey("getCorsTrustedOrigins", t, func() {
		Convey("When the env var is empty, it defaults to the local frontend", func() {
			t.Setenv("CORS_TRUSTED_ORIGINS", "")

			got, err := getCorsTrustedOrigins()

			So(err, ShouldBeNil)
			So(got, ShouldHaveLength, 1)
			So(got, ShouldContainKey, "http://localhost:3000")
		})

		Convey("When several origins are given, it trims and keeps each", func() {
			t.Setenv("CORS_TRUSTED_ORIGINS", " https://app.example.com , http://localhost:3000 ,")

			got, err := getCorsTrustedOrigins()

			So(err, ShouldBeNil)
			So(got, ShouldHaveLength, 2)
			So(got, ShouldContainKey, "https://app.example.com")
			So(got, ShouldContainKey, "http://localhost:3000")
		})

		for _, bad := range []string{"*", "example.com", "://broken", "https://ok.example.com,*"} {
			Convey("When an origin is invalid ("+bad+"), it fails startup", func() {
				t.Setenv("CORS_TRUSTED_ORIGINS", bad)

				got, err := getCorsTrustedOrigins()

				So(got, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "invalid CORS origin")
			})
		}

		Convey("When only separators are given, it requires at least one origin", func() {
			t.Setenv("CORS_TRUSTED_ORIGINS", " , ,")

			_, err := getCorsTrustedOrigins()

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "at least one origin")
		})
	})
}

func TestGetServerTimeout(t *testing.T) {
	for _, key := range []string{"READ_HEADER_TIMEOUT", "READ_TIMEOUT", "WRITE_TIMEOUT", "IDLE_TIMEOUT", "REQUEST_TIMEOUT"} {
		t.Setenv(key, "")
	}

	Convey("getServerTimeout", t, func() {
		Convey("When nothing is set, it applies the defaults", func() {
			cfg := &app.Config{}

			So(getServerTimeout(cfg), ShouldBeNil)
			So(cfg.Timeouts.ReadHeaderTimeout, ShouldEqual, 5*time.Second)
			So(cfg.Timeouts.ReadTimeout, ShouldEqual, 10*time.Second)
			So(cfg.Timeouts.WriteTimeout, ShouldEqual, 30*time.Second)
			So(cfg.Timeouts.IdleTimeout, ShouldEqual, 60*time.Second)
			So(cfg.Timeouts.RequestTimeout, ShouldEqual, 30*time.Second)
		})

		Convey("When a timeout is overridden, it is used", func() {
			t.Setenv("READ_TIMEOUT", "3s")
			cfg := &app.Config{}

			So(getServerTimeout(cfg), ShouldBeNil)
			So(cfg.Timeouts.ReadTimeout, ShouldEqual, 3*time.Second)
		})

		Convey("When a timeout is not positive, it fails and names the variable", func() {
			t.Setenv("IDLE_TIMEOUT", "-1s")

			err := getServerTimeout(&app.Config{})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "IDLE_TIMEOUT must be positive")
		})
	})
}
