package logrus

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1"

	"github.com/google/wire"
	"github.com/sirupsen/logrus"
)

var (
	// Providers is what we offer to external implementers
	Providers = wire.NewSet(
		ProvideLogrus,
		ProvideLogger,
	)
)

var _ logging.Logger = (*Logger)(nil)

// Logger is our log wrapper
type Logger struct {
	logger *logrus.Logger
}

// ProvideLogrus builds a logrus logger to our specs
func ProvideLogrus(debug bool) *logrus.Logger {
	logger := logrus.New()
	if debug {
		logger.SetLevel(logrus.DebugLevel)
	}

	logger.SetFormatter(&logrus.JSONFormatter{
		DataKey: "meta",
		// PrettyPrint: true,
	})

	logger.SetReportCaller(true)

	return logger
}

// ProvideLogger builds a new logger
func ProvideLogger(logger *logrus.Logger) logging.Logger {
	l := &Logger{logger: logger}
	return l
}

// Info satisfies our contract for the logging.Logger Info method.
func (l *Logger) Info(input string) {
	l.logger.Infoln(input)
}

// Debug satisfies our contract for the logging.Logger Debug method.
func (l *Logger) Debug(input string) {
	l.logger.Debugln(input)
}

// Error satisfies our contract for the logging.Logger Error method.
func (l *Logger) Error(input string) {
	l.logger.Errorln(input)
}

// WithValues satisfies our contract for the logging.Logger WithValues method.
func (l *Logger) WithValues(values map[string]interface{}) logging.Logger {
	return &entryWrapper{l.logger.WithFields(values)}
}

// WithValue satisfies our contract for the logging.Logger WithValue method.
func (l *Logger) WithValue(key string, value interface{}) logging.Logger {
	return &entryWrapper{l.logger.WithField(key, value)}
}

// WithError satisfies our contract for the logging.Logger WithError method.
func (l *Logger) WithError(err error) logging.Logger {
	return l.WithError(err)
}

type entryWrapper struct {
	*logrus.Entry
}

// Info satisfies our contract for the logging.Logger Info method.
func (e *entryWrapper) Info(input string) {
	e.Infoln(input)
}

// Debug satisfies our contract for the logging.Logger Debug method.
func (e *entryWrapper) Debug(input string) {
	e.Debugln(input)
}

// Error satisfies our contract for the logging.Logger Error method.
func (e *entryWrapper) Error(input string) {
	e.Errorln(input)
}

// WithValues satisfies our contract for the logging.Logger WithValues method.
func (e *entryWrapper) WithValues(values map[string]interface{}) logging.Logger {
	return &entryWrapper{e.WithFields(values)}
}

// WithValue satisfies our contract for the logging.Logger WithValue method.
func (e *entryWrapper) WithValue(key string, value interface{}) logging.Logger {
	return &entryWrapper{e.WithField(key, value)}
}

// WithError satisfies our contract for the logging.Logger WithError method.
func (e *entryWrapper) WithError(err error) logging.Logger {
	return e.WithError(err)
}
