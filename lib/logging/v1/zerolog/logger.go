package zerolog

import (
	"net/http"
	"os"

	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1"

	"github.com/google/wire"
	"github.com/rs/zerolog"
)

var (
	// Providers is what we offer to external implementers
	Providers = wire.NewSet(
		ProvideZerologger,
		ProvideLogger,
	)
)

var _ logging.Logger = (*Logger)(nil)

// Logger is our log wrapper
type Logger struct {
	logger zerolog.Logger
}

// ProvideZerologger builds a new zerologger
func ProvideZerologger() zerolog.Logger {
	l := zerolog.New(os.Stdout)
	return l
}

// ProvideLogger builds a new logger
func ProvideLogger(logger zerolog.Logger) logging.Logger {
	l := &Logger{logger: logger}
	return l
}

// SetLevel sets the log level for our logger
func (l *Logger) SetLevel(level logging.Level) {
	var lvl zerolog.Level
	switch level {
	case logging.InfoLevel:
		lvl = zerolog.InfoLevel
	case logging.DebugLevel:
		lvl = zerolog.DebugLevel
	case logging.ErrorLevel:
		lvl = zerolog.ErrorLevel
	}
	l.logger = l.logger.Level(lvl)
}

// Info satisfies our contract for the logging.Logger Info method.
func (l *Logger) Info(input string) {
	l.logger.Info().Msg(input)
}

// Debug satisfies our contract for the logging.Logger Debug method.
func (l *Logger) Debug(input string) {
	l.logger.Debug().Msg(input)
}

// Error satisfies our contract for the logging.Logger Error method.
func (l *Logger) Error(err error, input string) {
	l.logger.Error().Err(err).Msg(input)
}

// Fatal satisfies our contract for the logging.Logger Fatal method.
func (l *Logger) Fatal(err error) {
	l.logger.Fatal().Err(err).Msg("")
}

// Print satisfies our contract for the logging.Logger Print method.
func (l *Logger) Print(input ...interface{}) {
	l.logger.Print(input...)
}

// WithValues satisfies our contract for the logging.Logger WithValues method.
func (l *Logger) WithValues(values map[string]interface{}) logging.Logger {
	var l2 zerolog.Logger

	for key, val := range values {
		l2 = l.logger.With().Interface(key, val).Logger()
	}

	return &Logger{logger: l2}
}

// WithValue satisfies our contract for the logging.Logger WithValue method.
func (l *Logger) WithValue(key string, value interface{}) logging.Logger {
	l2 := l.logger.With().Interface(key, value).Logger()
	return &Logger{logger: l2}
}

// WithError satisfies our contract for the logging.Logger WithError method.
func (l *Logger) WithError(err error) logging.Logger {
	l2 := l.logger.With().Err(err).Logger()
	return &Logger{logger: l2}
}

// WithRequest satisfies our contract for the logging.Logger WithRequest method.
func (l *Logger) WithRequest(req *http.Request) logging.Logger {
	l2 := l.logger.With().
		Str("path", req.URL.Path).
		Str("method", req.Method).
		Str("query", req.URL.RawQuery).
		Logger()
	return &Logger{logger: l2}
}
