package main

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func setValidDBEnv(t *testing.T) {
	t.Helper()
	t.Setenv("DB_NAME", "learnflow")
	t.Setenv("DB_USER", "learnflow")
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PASSWORD", "secret")
}

func setValidSMTPEnv(t *testing.T) {
	t.Helper()
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_USERNAME", "user")
	t.Setenv("SMTP_PASSWORD", "pass")
	t.Setenv("SMTP_SENDER", "noreply@example.com")
}

func TestGetMailerConfig(t *testing.T) {
	Convey("getMailerConfig", t, func() {
		Convey("When all SMTP env vars are set, it succeeds", func() {
			setValidSMTPEnv(t)
			cfg := &Config{}

			err := getMailerConfig(cfg, "production")

			So(err, ShouldBeNil)
			So(cfg.SMTP.Host, ShouldEqual, "smtp.example.com")
			So(cfg.SMTP.Port, ShouldEqual, 587)
		})

		Convey("When SMTP_HOST is unset in production, it errors", func() {
			t.Setenv("SMTP_HOST", "")
			cfg := &Config{}

			err := getMailerConfig(cfg, "production")

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "smtp host")
		})

		Convey("When SMTP_HOST is unset outside production, it falls back to the stub host", func() {
			t.Setenv("SMTP_HOST", "")
			t.Setenv("SMTP_USERNAME", "user")
			t.Setenv("SMTP_PASSWORD", "pass")
			t.Setenv("SMTP_SENDER", "noreply@example.com")
			cfg := &Config{}

			err := getMailerConfig(cfg, "development")

			So(err, ShouldBeNil)
			So(cfg.SMTP.Host, ShouldEqual, "stub")
		})

		Convey("When SMTP_USERNAME is missing, it errors", func() {
			setValidSMTPEnv(t)
			t.Setenv("SMTP_USERNAME", "")
			cfg := &Config{}

			err := getMailerConfig(cfg, "production")

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "smtp username")
		})

		Convey("When SMTP_PASSWORD is missing, it errors", func() {
			setValidSMTPEnv(t)
			t.Setenv("SMTP_PASSWORD", "")
			cfg := &Config{}

			err := getMailerConfig(cfg, "production")

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "smtp password")
		})

		Convey("When SMTP_SENDER is missing, it errors", func() {
			setValidSMTPEnv(t)
			t.Setenv("SMTP_SENDER", "")
			cfg := &Config{}

			err := getMailerConfig(cfg, "production")

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "smtp sender")
		})
	})
}

func TestGetAppConfig(t *testing.T) {
	Convey("getAppConfig", t, func() {
		Convey("When DB and SMTP env vars are all valid, it succeeds", func() {
			setValidDBEnv(t)
			setValidSMTPEnv(t)

			cfg, err := getAppConfig("production")

			So(err, ShouldBeNil)
			So(cfg.Env, ShouldEqual, "production")
			So(cfg.Database.DSN, ShouldContainSubstring, "learnflow")
		})

		Convey("When DB config is invalid (missing DB_NAME), it errors before touching SMTP", func() {
			t.Setenv("DB_NAME", "")
			t.Setenv("DB_USER", "learnflow")
			t.Setenv("DB_HOST", "localhost")
			t.Setenv("DB_PASSWORD", "secret")

			_, err := getAppConfig("production")

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "DB_NAME")
		})

		Convey("When DB config is valid but SMTP is invalid, it errors", func() {
			setValidDBEnv(t)
			t.Setenv("SMTP_HOST", "")

			_, err := getAppConfig("production")

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "smtp")
		})
	})
}
