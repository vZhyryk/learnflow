// Package logger provides a structured JSON logger with sanitization and configurable trace levels.
package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"learnflow_backend/internal/infrastructure/convert"
	"learnflow_backend/internal/infrastructure/sanitizer"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

// Level represents log severity.
type Level int8

// Log level constants ordered from least to most severe.
const (
	LevelInfo  Level = iota // informational messages
	LevelError              // non-fatal errors
	LevelFatal              // fatal errors — process exits after logging
)

// String returns the human-readable name of the log level.
func (l Level) String() string {
	switch l {
	case LevelInfo:
		return "INFO"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return ""
	}
}

// Logger writes structured JSON log entries to an io.Writer.
// All writes are protected by a mutex — safe for concurrent use.
type Logger struct {
	out        io.Writer
	mu         sync.Mutex
	sanitizer  *sanitizer.Sanitizer
	traceLevel Level // minimum level at which a goroutine stack trace is added
}

// New creates a Logger. A nil sanit is a no-op passthrough — fine for dev/test, never in prod.
// traceLevel controls when stack traces are attached (LevelError: dev, LevelFatal: prod).
func New(out io.Writer, sanit *sanitizer.Sanitizer, traceLevel Level) *Logger {
	if sanit == nil {
		sanit = sanitizer.NewSanitizer("", 0, nil)
	}
	return &Logger{
		out:        out,
		sanitizer:  sanit,
		traceLevel: traceLevel,
	}
}

// Info logs an informational message with optional key-value properties.
func (l *Logger) Info(message string, properties map[string]any) {
	err := l.log(LevelInfo, message, properties)
	if err != nil {
		panic(fmt.Errorf("logger.Info: %w", err))
	}
}

// Error logs an error. No-op if err is nil.
func (l *Logger) Error(err error, properties map[string]any) {
	if err != nil {
		logErr := l.log(LevelError, err.Error(), properties)
		if logErr != nil {
			panic(fmt.Errorf("logger.Error: %w", logErr))
		}
	}
}

// Fatal logs a fatal error and exits with code 1.
// Panics if err is nil — calling Fatal(nil) is a programmer error (the caller has a bug).
func (l *Logger) Fatal(err error, properties map[string]any) {
	if err == nil {
		panic("logger.Fatal called with nil error")
	}
	err = l.log(LevelFatal, err.Error(), properties)
	if err != nil {
		panic(fmt.Errorf("logger.Fatal: %w", err))
	}
	os.Exit(1)
}

func (l *Logger) log(level Level, message string, properties map[string]any) error {
	aux := struct {
		Level      string         `json:"level"`
		Time       string         `json:"time"`
		Message    string         `json:"message"`
		Properties map[string]any `json:"properties,omitempty"`
		Trace      string         `json:"trace,omitempty"`
	}{
		Level: level.String(),
		Time:  time.Now().UTC().Format(time.RFC3339),
	}

	aux.Message = l.sanitizer.SanitizeString(message)

	sanitized := l.sanitizer.Sanitize(properties)
	if props, ok := convert.ToMapStringAny(sanitized); ok {
		aux.Properties = props
	}

	if level >= l.traceLevel {
		aux.Trace = l.sanitizer.SanitizeString(string(debug.Stack()))
	}

	line, err := json.Marshal(aux)
	if err != nil {
		line = []byte(LevelError.String() + ": unable to marshal log message: " + err.Error())
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	_, err = l.out.Write(append(line, '\n'))
	if err != nil {
		return fmt.Errorf("logger: write: %w", err)
	}
	return nil
}

// Write implements io.Writer so Logger can be used as http.Server.ErrorLog.
// The raw bytes are treated as an error-level message.
func (l *Logger) Write(message []byte) (int, error) {
	if err := l.log(LevelError, string(message), nil); err != nil {
		return 0, err
	}
	return len(message), nil
}
