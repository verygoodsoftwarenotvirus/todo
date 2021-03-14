package logging

import (
	"net/http"
	"strings"
	"time"

	chimiddleware "github.com/go-chi/chi/middleware"
)

const (
	// LoggerNameKey is a key we can use to denote logger names across implementations.
	LoggerNameKey = "__name__"
)

type (
	level int

	// Level is a simple string alias for dependency injection's sake.
	Level *level

	// RequestIDFunc fetches a string ID from a request.
	RequestIDFunc func(*http.Request) string
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
	SetRequestIDFunc(RequestIDFunc)

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

var doNotLog = map[string]struct{}{
	// "/metrics": {},
	"/build/":  {},
	"/assets/": {},
}

// BuildLoggingMiddleware builds a logging middleware.
func BuildLoggingMiddleware(logger Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			ww := chimiddleware.NewWrapResponseWriter(res, req.ProtoMajor)

			start := time.Now()
			next.ServeHTTP(ww, req)

			shouldLog := true
			for route := range doNotLog {
				if strings.HasPrefix(req.URL.Path, route) || req.URL.Path == route {
					shouldLog = false
					break
				}
			}

			if shouldLog {
				logger.WithRequest(req).WithValues(map[string]interface{}{
					"status":  ww.Status(),
					"elapsed": time.Since(start).Milliseconds(),
					"written": ww.BytesWritten(),
				}).Debug("response served")
			}
		})
	}
}
