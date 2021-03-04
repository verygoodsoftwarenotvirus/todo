package zerolog

import (
	"net/http"
	"os"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"

	"github.com/rs/zerolog"
)

func init() {
	zerolog.CallerSkipFrameCount++
}

// Logger is our log wrapper.
type Logger struct {
	logger        zerolog.Logger
	requestIDFunc logging.RequestIDFunc
}

// buildZerologger builds a new zerologger.
func buildZerologger() zerolog.Logger {
	return zerolog.
		New(os.Stdout).
		With().
		Timestamp().
		Logger()
}

// NewLogger builds a new logger.
func NewLogger() logging.Logger {
	return &Logger{logger: buildZerologger()}
}

// NewLoggerWithSource builds a new logger.
func NewLoggerWithSource(logger *zerolog.Logger) logging.Logger {
	return &Logger{logger: *logger}
}

// WithName is our obligatory contract fulfillment function.
// Zerolog doesn't support named loggers :( so we have this workaround.
func (l *Logger) WithName(name string) logging.Logger {
	l2 := l.logger.With().Str(logging.LoggerNameKey, name).Logger()
	return &Logger{logger: l2}
}

// SetLevel sets the log level for our logger.
func (l *Logger) SetLevel(level logging.Level) {
	var lvl zerolog.Level

	switch level {
	case logging.InfoLevel:
		lvl = zerolog.InfoLevel
	case logging.DebugLevel:
		l.logger = l.logger.With().Logger()
		lvl = zerolog.DebugLevel
	case logging.WarnLevel:
		l.logger = l.logger.With().Caller().Logger()
		lvl = zerolog.WarnLevel
	case logging.ErrorLevel:
		l.logger = l.logger.With().Caller().Logger()
		lvl = zerolog.ErrorLevel
	}

	l.logger = l.logger.Level(lvl)
}

// SetRequestIDFunc sets the request ID retrieval function.
func (l *Logger) SetRequestIDFunc(f logging.RequestIDFunc) {
	if f != nil {
		l.requestIDFunc = f
	}
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
	l.logger.Error().Caller().Err(err).Msg(input)
}

// Fatal satisfies our contract for the logging.Logger Fatal method.
func (l *Logger) Fatal(err error) {
	l.logger.Fatal().Caller().Err(err).Msg("")
}

// Printf satisfies our contract for the logging.Logger Printf method.
func (l *Logger) Printf(format string, args ...interface{}) {
	l.logger.Printf(format, args...)
}

// WithValues satisfies our contract for the logging.Logger WithValues method.
func (l *Logger) WithValues(values map[string]interface{}) logging.Logger {
	var l2 = l.logger.With().Logger()

	for key, val := range values {
		l2 = l2.With().Interface(key, val).Logger()
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
		Logger()

	if req.URL.RawQuery != "" {
		l2 = l2.With().Str("query", req.URL.RawQuery).Logger()
	}

	if l.requestIDFunc != nil {
		if reqID := l.requestIDFunc(req); reqID != "" {
			l2 = l2.With().Str("request_id", reqID).Logger()
		}
	}

	return &Logger{logger: l2}
}
