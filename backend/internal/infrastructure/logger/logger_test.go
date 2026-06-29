package logger_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"learnflow_backend/internal/infrastructure/logger"
	"learnflow_backend/internal/infrastructure/sanitizer"

	. "github.com/smartystreets/goconvey/convey"
)

func newLogger(buf *bytes.Buffer) *logger.Logger {
	s := sanitizer.NewSanitizer("***", 2000, nil)
	return logger.New(buf, s, logger.LevelFatal)
}

func parseLog(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()
	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse log output: %v", err)
	}
	return entry
}

// --- Level.String ---

func TestLevelString(t *testing.T) {
	Convey("Level.String", t, func() {
		So(logger.LevelInfo.String(), ShouldEqual, "INFO")
		So(logger.LevelError.String(), ShouldEqual, "ERROR")
		So(logger.LevelFatal.String(), ShouldEqual, "FATAL")
	})
}

// --- Info ---

func TestLoggerInfo(t *testing.T) {
	Convey("Logger.Info", t, func() {
		var buf bytes.Buffer
		l := newLogger(&buf)

		Convey("writes JSON entry with level INFO and message", func() {
			l.Info("test message", nil)
			entry := parseLog(t, &buf)
			So(entry["level"], ShouldEqual, "INFO")
			So(entry["message"], ShouldEqual, "test message")
			So(entry["time"], ShouldNotBeEmpty)
		})

		Convey("includes properties when provided", func() {
			l.Info("with props", map[string]any{"user_id": "u-1"})
			entry := parseLog(t, &buf)
			props, ok := entry["properties"].(map[string]any)
			So(ok, ShouldBeTrue)
			So(props["user_id"], ShouldEqual, "u-1")
		})

		Convey("redacts sensitive properties", func() {
			l.Info("login", map[string]any{"password": "s3cr3t"})
			entry := parseLog(t, &buf)
			props, ok := entry["properties"].(map[string]any)
			So(ok, ShouldBeTrue)
			So(props["password"], ShouldEqual, "***")
		})

		Convey("does not include trace when traceLevel is Fatal", func() {
			l.Info("no trace", nil)
			entry := parseLog(t, &buf)
			_, hasTrace := entry["trace"]
			So(hasTrace, ShouldBeFalse)
		})
	})
}

// --- Error ---

func TestLoggerError(t *testing.T) {
	Convey("Logger.Error", t, func() {
		var buf bytes.Buffer
		l := newLogger(&buf)

		Convey("When err is nil", func() {
			l.Error(nil, nil)
			Convey("writes nothing", func() {
				So(buf.Len(), ShouldEqual, 0)
			})
		})

		Convey("When err is non-nil", func() {
			l.Error(errors.New("something broke"), nil)
			Convey("writes JSON entry with level ERROR", func() {
				entry := parseLog(t, &buf)
				So(entry["level"], ShouldEqual, "ERROR")
				So(entry["message"], ShouldEqual, "something broke")
			})
		})
	})
}

// --- Write (io.Writer) ---

func TestLoggerWrite(t *testing.T) {
	Convey("Logger.Write", t, func() {
		var buf bytes.Buffer
		l := newLogger(&buf)

		n, err := l.Write([]byte("raw message from http.Server"))
		Convey("Then writes ERROR entry and returns byte count", func() {
			So(err, ShouldBeNil)
			So(n, ShouldEqual, len("raw message from http.Server"))
			entry := parseLog(t, &buf)
			So(entry["level"], ShouldEqual, "ERROR")
			So(entry["message"], ShouldContainSubstring, "raw message")
		})
	})
}

// --- Fatal with nil panics ---

func TestLoggerFatalNilPanics(t *testing.T) {
	Convey("Logger.Fatal(nil) panics", t, func() {
		var buf bytes.Buffer
		l := newLogger(&buf)
		So(func() { l.Fatal(nil, nil) }, ShouldPanic)
	})
}

// --- trace added at traceLevel ---

func TestLoggerTraceLevel(t *testing.T) {
	Convey("Logger with traceLevel=LevelInfo", t, func() {
		var buf bytes.Buffer
		s := sanitizer.NewSanitizer("***", 2000, nil)
		l := logger.New(&buf, s, logger.LevelInfo)

		l.Info("with trace", nil)
		Convey("Then trace is included in output", func() {
			entry := parseLog(t, &buf)
			So(entry["trace"], ShouldNotBeEmpty)
		})
	})
}
