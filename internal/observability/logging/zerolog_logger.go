package logging

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
)

const here = "gitlab.com/verygoodsoftwarenotvirus/todo/"

func init() {
	zerolog.CallerSkipFrameCount += 2
	zerolog.DisableSampling(true)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	zerolog.TimestampFunc = func() time.Time {
		return time.Now().UTC()
	}
	zerolog.CallerMarshalFunc = func(file string, line int) string {
		return strings.TrimPrefix(file, here) + ":" + strconv.Itoa(line)
	}
}

// logger is our log wrapper.
type zerologLogger struct {
	requestIDFunc RequestIDFunc
	logger        zerolog.Logger
}

// buildZerologger builds a new zerologger.
func buildZerologger() zerolog.Logger {
	return zerolog.New(os.Stdout).With().Timestamp().Logger().Level(zerolog.InfoLevel)
}

// NewZerologLogger builds a new zerologLogger.
func NewZerologLogger() Logger {
	return &zerologLogger{logger: buildZerologger()}
}

// WithName is our obligatory contract fulfillment function.
// Zerolog doesn't support named loggers :( so we have this workaround.
func (l *zerologLogger) WithName(name string) Logger {
	l2 := l.logger.With().Str(LoggerNameKey, name).Logger()
	return &zerologLogger{logger: l2}
}

// SetLevel sets the log level for our zerologLogger.
func (l *zerologLogger) SetLevel(level Level) {
	var lvl zerolog.Level

	switch level {
	case InfoLevel:
		lvl = zerolog.InfoLevel
	case DebugLevel:
		l.logger = l.logger.With().Logger()
		lvl = zerolog.DebugLevel
	case WarnLevel:
		l.logger = l.logger.With().Caller().Logger()
		lvl = zerolog.WarnLevel
	case ErrorLevel:
		l.logger = l.logger.With().Caller().Logger()
		lvl = zerolog.ErrorLevel
	}

	l.logger = l.logger.Level(lvl)
}

// SetRequestIDFunc sets the request ID retrieval function.
func (l *zerologLogger) SetRequestIDFunc(f RequestIDFunc) {
	if f != nil {
		l.requestIDFunc = f
	}
}

// Info satisfies our contract for the logging.Logger Info method.
func (l *zerologLogger) Info(input string) {
	l.logger.Info().Msg(input)
}

// Debug satisfies our contract for the logging.Logger Debug method.
func (l *zerologLogger) Debug(input string) {
	l.logger.Debug().Msg(input)
}

// Error satisfies our contract for the logging.Logger Error method.
func (l *zerologLogger) Error(err error, input string) {
	if err == nil {
		l.logger.Error().Stack().Caller().Str("err", "nil").Msg(input)
	} else {
		l.logger.Error().Stack().Caller().Err(err).Msg(input)
	}
}

// Fatal satisfies our contract for the logging.Logger Fatal method.
func (l *zerologLogger) Fatal(err error) {
	l.logger.Fatal().Caller().Err(err).Msg(err.Error())
}

// Printf satisfies our contract for the logging.Logger Printf method.
func (l *zerologLogger) Printf(format string, args ...interface{}) {
	l.logger.Printf(format, args...)
}

// Clone satisfies our contract for the logging.Logger WithValue method.
func (l *zerologLogger) Clone() Logger {
	l2 := l.logger.With().Logger()
	return &zerologLogger{logger: l2}
}

// WithValue satisfies our contract for the logging.Logger WithValue method.
func (l *zerologLogger) WithValue(key string, value interface{}) Logger {
	l2 := l.logger.With().Interface(key, value).Logger()
	return &zerologLogger{logger: l2}
}

// WithValues satisfies our contract for the logging.Logger WithValues method.
func (l *zerologLogger) WithValues(values map[string]interface{}) Logger {
	var l2 = l.logger.With().Logger()

	for key, val := range values {
		l2 = l2.With().Interface(key, val).Logger()
	}

	return &zerologLogger{logger: l2}
}

// WithError satisfies our contract for the logging.Logger WithError method.
func (l *zerologLogger) WithError(err error) Logger {
	l2 := l.logger.With().Err(err).Logger()
	return &zerologLogger{logger: l2}
}

func (l *zerologLogger) attachRequestToLog(req *http.Request) zerolog.Logger {
	if req != nil {
		l2 := l.logger.With().
			Str("method", req.Method).
			Logger()

		if req.URL != nil {
			l2 = l2.With().Str("path", req.URL.Path).Logger()
			if req.URL.RawQuery != "" {
				l2 = l2.With().Str("query", req.URL.RawQuery).Logger()
			}
		}

		if l.requestIDFunc != nil {
			if reqID := l.requestIDFunc(req); reqID != "" {
				l2 = l2.With().Str("request.id", reqID).Logger()
			}
		}

		return l2
	}

	return l.logger
}

// WithRequest satisfies our contract for the logging.Logger WithRequest method.
func (l *zerologLogger) WithRequest(req *http.Request) Logger {
	return &zerologLogger{logger: l.attachRequestToLog(req)}
}

// WithResponse satisfies our contract for the logging.Logger WithResponse method.
func (l *zerologLogger) WithResponse(res *http.Response) Logger {
	l2 := l.logger.With().Logger()
	if res != nil {
		l2 = l.attachRequestToLog(res.Request).With().Int(keys.ResponseStatusKey, res.StatusCode).Logger()
	}

	return &zerologLogger{logger: l2}
}
