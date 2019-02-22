package logrus

import (
	"net/http"

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

// WithName is our obligatory contract fulfillment function
// Logrus doesn't support named loggers :(
func (l *Logger) WithName(name string) logging.Logger {
	return l
}

// SetLevel sets the log level
func (l *Logger) SetLevel(level logging.Level) {
	var lvl logrus.Level
	switch level {
	case logging.InfoLevel:
		lvl = logrus.InfoLevel
	case logging.DebugLevel:
		lvl = logrus.DebugLevel
	case logging.ErrorLevel:
		lvl = logrus.ErrorLevel
	}
	l.logger.SetLevel(lvl)
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
func (l *Logger) Error(err error, input string) {
	l.logger.WithField("err", err).Errorln(input)
}

// Fatal satisfies our contract for the logging.Logger Fatal method.
func (l *Logger) Fatal(err error) {
	l.logger.WithField("err", err).Fatal()
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

// WithRequest satisfies our contract for the logging.Logger WithRequest method.
func (l *Logger) WithRequest(req *http.Request) logging.Logger {
	return &entryWrapper{l.logger.WithFields(map[string]interface{}{
		"path":   req.URL.Path,
		"method": req.Method,
		"query":  req.URL.RawQuery,
	})}
}
