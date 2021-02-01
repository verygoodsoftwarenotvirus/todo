package logging

import (
	"net/http"
)

const (
	// LoggerNameKey is a key we can use to denote logger names across implementations.
	LoggerNameKey = "__name__"
)

type (
	level int

	// Level is a simple string alias for dependency injection's sake.
	Level *level
)

var (
	// InfoLevel describes a info-level log.
	InfoLevel Level = new(level)
	// DebugLevel describes a debug-level log.
	DebugLevel Level = new(level)
	// ErrorLevel describes a error-level log.
	ErrorLevel Level = new(level)
	// WarnLevel describes a warn-level log.
	WarnLevel Level = new(level)
)

// Logger represents a simple logging interface we can build wrappers around.
type Logger interface {
	Info(string)
	Debug(string)
	Error(error, string)
	Fatal(error)
	Printf(string, ...interface{})

	SetLevel(Level)

	// Builder functions
	WithName(string) Logger
	WithValues(map[string]interface{}) Logger
	WithValue(string, interface{}) Logger
	WithRequest(*http.Request) Logger
	WithError(error) Logger
}

// EnsureLogger guarantees that a logger is available.
func EnsureLogger(logger Logger) Logger {
	if logger != nil {
		return logger
	}

	return NewNonOperationalLogger()
}
