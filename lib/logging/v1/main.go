package logging

import (
	"net/http"
)

type (
	// Level is a simple string alias for dependency injection's sake
	Level string

	// LoggerName is a simple string alias for dependency injection's sake
	LoggerName string
)

var (
	// InfoLevel describes a info-level log
	InfoLevel Level = "info"
	// DebugLevel describes a debug-level log
	DebugLevel Level = "debug"
	// ErrorLevel describes a error-level log
	ErrorLevel Level = "error"
	// WarnLevel describes a warn-level log
	WarnLevel Level = "warn"
)

// Logger represents a simple logging interface we can build wrappers around.
type Logger interface {
	Info(string)
	Debug(string)
	Error(error, string)
	Fatal(error)

	SetLevel(Level)

	// Builder funcs
	WithName(string) Logger
	WithValues(map[string]interface{}) Logger
	WithValue(string, interface{}) Logger
	WithRequest(*http.Request) Logger
	WithError(error) Logger
}
