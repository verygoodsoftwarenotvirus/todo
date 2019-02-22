package zap

import (
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1"

	"github.com/google/wire"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Providers is what we offer to external implementers
	Providers = wire.NewSet(
		ProvideZapLogger,
		ProvideLogger,
	)

	_ logging.Logger = (*Logger)(nil)
)

// Logger is our log wrapper
type Logger struct {
	logger *zap.Logger
}

// ProvideZapLogger builds a new zap logger
func ProvideZapLogger(name logging.LoggerName) (*zap.Logger, error) {
	config := zap.NewDevelopmentConfig()
	logger, err := config.Build()
	logger = logger.Named(string(name))
	return logger, err
}

// ProvideLogger builds a new logger
func ProvideLogger(logger *zap.Logger) logging.Logger {
	l := &Logger{logger: logger}
	return l
}

// WithName is our obligatory contract fulfillment function
// Zerolog doesn't support named loggers :(
func (l *Logger) WithName(name string) logging.Logger {
	return &Logger{logger: l.logger.Named(name)}
}

// Info satisfies our contract for the logging.Logger Info method.
func (l *Logger) Info(input string) {
	l.logger.Info(input)
}

// Debug satisfies our contract for the logging.Logger Debug method.
func (l *Logger) Debug(input string) {
	l.logger.Debug(input)
}

// Error satisfies our contract for the logging.Logger Error method.
func (l *Logger) Error(err error, input string) {
	l.logger.Error(input)
}

// Fatal satisfies our contract for the logging.Logger Fatal method.
func (l *Logger) Fatal(err error) {
	l.logger.Fatal(err.Error())
}

// SetLevel satisfies our contract for the logging.Logger SetLevel method.
// there doesn't seem to be a zap equivalent for this
func (l *Logger) SetLevel(level logging.Level) {

}

// WithValues satisfies our contract for the logging.Logger WithValues method.
func (l *Logger) WithValues(values map[string]interface{}) logging.Logger {
	var l2 *zap.Logger

	fields := []zapcore.Field{}
	for key, value := range values {
		fields = append(fields, zap.Any(key, value))
	}

	l2 = l.logger.With(fields...)
	return &Logger{logger: l2}
}

// WithValue satisfies our contract for the logging.Logger WithValue method.
func (l *Logger) WithValue(key string, value interface{}) logging.Logger {
	l2 := l.logger.With(zapcore.Field{Key: key, Interface: value})
	return &Logger{logger: l2}
}

// WithError satisfies our contract for the logging.Logger WithError method.
func (l *Logger) WithError(err error) logging.Logger {
	l2 := l.logger.With(zap.Error(err))
	return &Logger{logger: l2}
}

// WithRequest satisfies our contract for the logging.Logger WithRequest method.
func (l *Logger) WithRequest(req *http.Request) logging.Logger {
	l2 := l.logger.With(
		zapcore.Field{Key: "path", String: req.URL.Path},
		zapcore.Field{Key: "method", String: req.Method},
		zapcore.Field{Key: "query", String: req.URL.RawQuery},
	)

	return &Logger{logger: l2}
}
