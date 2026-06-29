package env_test

import (
	"os"
	"testing"
	"time"

	"learnflow_backend/internal/infrastructure/env"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetStringEnv(t *testing.T) {
	const key = "LEARNFLOW_TEST_STRING"

	Convey("GetStringEnv", t, func() {
		Reset(func() { So(os.Unsetenv(key), ShouldBeNil) })

		Convey("When variable is unset", func() {
			Convey("returns fallback", func() {
				So(env.GetStringEnv(key, "default"), ShouldEqual, "default")
			})
		})

		Convey("When variable is set", func() {
			So(os.Setenv(key, "hello"), ShouldBeNil)
			Convey("returns the value", func() {
				So(env.GetStringEnv(key, "default"), ShouldEqual, "hello")
			})
		})

		Convey("When variable is blank", func() {
			So(os.Setenv(key, "   "), ShouldBeNil)
			Convey("returns fallback", func() {
				So(env.GetStringEnv(key, "default"), ShouldEqual, "default")
			})
		})

		Convey("When variable has surrounding whitespace", func() {
			So(os.Setenv(key, "  trimmed  "), ShouldBeNil)
			Convey("returns trimmed value", func() {
				So(env.GetStringEnv(key, "default"), ShouldEqual, "trimmed")
			})
		})
	})
}

func TestGetIntEnv(t *testing.T) {
	const key = "LEARNFLOW_TEST_INT"

	Convey("GetIntEnv", t, func() {
		Reset(func() { So(os.Unsetenv(key), ShouldBeNil) })

		Convey("When variable is unset", func() {
			Convey("returns fallback", func() {
				So(env.GetIntEnv(key, 99), ShouldEqual, 99)
			})
		})

		Convey("When variable is a valid integer", func() {
			So(os.Setenv(key, "42"), ShouldBeNil)
			Convey("returns parsed value", func() {
				So(env.GetIntEnv(key, 0), ShouldEqual, 42)
			})
		})

		Convey("When variable is not a number", func() {
			So(os.Setenv(key, "abc"), ShouldBeNil)
			Convey("returns fallback", func() {
				So(env.GetIntEnv(key, 7), ShouldEqual, 7)
			})
		})

		Convey("When variable has whitespace", func() {
			So(os.Setenv(key, "  10  "), ShouldBeNil)
			Convey("returns parsed value", func() {
				So(env.GetIntEnv(key, 0), ShouldEqual, 10)
			})
		})
	})
}

func TestGetDurationEnv(t *testing.T) {
	const key = "LEARNFLOW_TEST_DUR"

	Convey("GetDurationEnv", t, func() {
		Reset(func() { So(os.Unsetenv(key), ShouldBeNil) })

		Convey("When variable is unset", func() {
			Convey("returns fallback", func() {
				So(env.GetDurationEnv(key, 5*time.Second), ShouldEqual, 5*time.Second)
			})
		})

		Convey("When variable is a valid duration", func() {
			So(os.Setenv(key, "2m30s"), ShouldBeNil)
			Convey("returns parsed duration", func() {
				So(env.GetDurationEnv(key, 0), ShouldEqual, 150*time.Second)
			})
		})

		Convey("When variable is invalid", func() {
			So(os.Setenv(key, "notaduration"), ShouldBeNil)
			Convey("returns fallback", func() {
				So(env.GetDurationEnv(key, time.Minute), ShouldEqual, time.Minute)
			})
		})
	})
}

func TestGetFloat64Env(t *testing.T) {
	const key = "LEARNFLOW_TEST_FLOAT"

	Convey("GetFloat64Env", t, func() {
		Reset(func() { So(os.Unsetenv(key), ShouldBeNil) })

		Convey("When variable is unset", func() {
			Convey("returns fallback", func() {
				So(env.GetFloat64Env(key, 1.5), ShouldEqual, 1.5)
			})
		})

		Convey("When variable is a valid float", func() {
			So(os.Setenv(key, "3.14"), ShouldBeNil)
			Convey("returns parsed value", func() {
				So(env.GetFloat64Env(key, 0), ShouldAlmostEqual, 3.14)
			})
		})

		Convey("When variable is invalid", func() {
			So(os.Setenv(key, "xyz"), ShouldBeNil)
			Convey("returns fallback", func() {
				So(env.GetFloat64Env(key, 9.9), ShouldEqual, 9.9)
			})
		})
	})
}

func TestGetBoolEnv(t *testing.T) {
	const key = "LEARNFLOW_TEST_BOOL"

	Convey("GetBoolEnv", t, func() {
		Reset(func() { So(os.Unsetenv(key), ShouldBeNil) })

		Convey("When variable is unset", func() {
			Convey("returns fallback", func() {
				So(env.GetBoolEnv(key, true), ShouldBeTrue)
			})
		})

		truthy := []string{"true", "1", "yes", "TRUE", "YES"}
		for _, v := range truthy {
			Convey(v+" → true", func() {
				So(os.Setenv(key, v), ShouldBeNil)
				So(env.GetBoolEnv(key, false), ShouldBeTrue)
			})
		}

		falsy := []string{"false", "0", "no", "FALSE", "NO"}
		for _, v := range falsy {
			Convey(v+" → false", func() {
				So(os.Setenv(key, v), ShouldBeNil)
				So(env.GetBoolEnv(key, true), ShouldBeFalse)
			})
		}

		Convey("When variable is unrecognised", func() {
			So(os.Setenv(key, "maybe"), ShouldBeNil)
			Convey("returns fallback", func() {
				So(env.GetBoolEnv(key, true), ShouldBeTrue)
			})
		})
	})
}
