package zerolog

import (
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
)

func init() {
	zerolog.CallerSkipFrameCount++
	zerolog.DisableSampling(true)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	zerolog.TimestampFunc = func() time.Time {
		return time.Now().UTC()
	}
}

// logger is our log wrapper.
type logger struct {
	requestIDFunc logging.RequestIDFunc
	logger        zerolog.Logger
}

// buildZerologger builds a new zerologger.
func buildZerologger() zerolog.Logger {
	return zerolog.New(os.Stdout).With().Timestamp().Logger()
}

// NewLogger builds a new logger.
func NewLogger() logging.Logger {
	return &logger{logger: buildZerologger()}
}

// NewLoggerWithSource builds a new logger.
func NewLoggerWithSource(l *zerolog.Logger) logging.Logger {
	return &logger{logger: *l}
}

// WithName is our obligatory contract fulfillment function.
// Zerolog doesn't support named loggers :( so we have this workaround.
func (l *logger) WithName(name string) logging.Logger {
	l2 := l.logger.With().Str(logging.LoggerNameKey, name).Logger()
	return &logger{logger: l2}
}

// SetLevel sets the log level for our logger.
func (l *logger) SetLevel(level logging.Level) {
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
func (l *logger) SetRequestIDFunc(f logging.RequestIDFunc) {
	if f != nil {
		l.requestIDFunc = f
	}
}

// Info satisfies our contract for the logging.Logger Info method.
func (l *logger) Info(input string) {
	l.logger.Info().Msg(input)
}

// Debug satisfies our contract for the logging.Logger Debug method.
func (l *logger) Debug(input string) {
	l.logger.Debug().Msg(input)
}

// Error satisfies our contract for the logging.Logger Error method.
func (l *logger) Error(err error, input string) {
	l.logger.Error().Stack().Caller().Err(err).Msg(input)
}

// Fatal satisfies our contract for the logging.Logger Fatal method.
func (l *logger) Fatal(err error) {
	l.logger.Fatal().Caller().Err(err).Msg("")
}

// Printf satisfies our contract for the logging.Logger Printf method.
func (l *logger) Printf(format string, args ...interface{}) {
	l.logger.Printf(format, args...)
}

// WithValues satisfies our contract for the logging.Logger WithValues method.
func (l *logger) WithValues(values map[string]interface{}) logging.Logger {
	var l2 = l.logger.With().Logger()

	for key, val := range values {
		l2 = l2.With().Interface(key, val).Logger()
	}

	return &logger{logger: l2}
}

// WithValue satisfies our contract for the logging.Logger WithValue method.
func (l *logger) WithValue(key string, value interface{}) logging.Logger {
	l2 := l.logger.With().Interface(key, value).Logger()
	return &logger{logger: l2}
}

// WithError satisfies our contract for the logging.Logger WithError method.
func (l *logger) WithError(err error) logging.Logger {
	l2 := l.logger.With().Err(err).Logger()
	return &logger{logger: l2}
}

func (l *logger) attachRequestToLog(req *http.Request) zerolog.Logger {
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

	return l2
}

// WithRequest satisfies our contract for the logging.Logger WithRequest method.
func (l *logger) WithRequest(req *http.Request) logging.Logger {
	return &logger{logger: l.attachRequestToLog(req)}
}

// WithResponse satisfies our contract for the logging.Logger WithResponse method.
func (l *logger) WithResponse(res *http.Response) logging.Logger {
	l2 := l.attachRequestToLog(res.Request).With().Int(keys.ResponseStatusKey, res.StatusCode).Logger()

	return &logger{logger: l2}
}
