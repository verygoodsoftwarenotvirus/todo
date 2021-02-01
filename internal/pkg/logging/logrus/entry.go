package logrus

import (
	"fmt"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/logging"

	"github.com/sirupsen/logrus"
)

// entryWrapper has repeats of many functions for the sake of interface implementation.
type entryWrapper struct {
	entry *logrus.Entry
}

// SetLevel is a noop used for the sake of interface implementation
// if you're calling this on an entry wrapper, then the level has
// already been determined. I have no use for a function which
// overrides that behavior, simple though it may be to paste it here.
func (e *entryWrapper) SetLevel(level logging.Level) {
	switch level {
	case logging.InfoLevel:
		e.entry.Level = logrus.InfoLevel
	case logging.DebugLevel:
		e.entry.Level = logrus.DebugLevel
	case logging.ErrorLevel:
		e.entry.Level = logrus.ErrorLevel
	case logging.WarnLevel:
		e.entry.Level = logrus.WarnLevel
	}
}

// WithName is our obligatory contract fulfillment function. Logrus doesn't support named loggers.
func (e *entryWrapper) WithName(name string) logging.Logger {
	e.entry = e.entry.WithField(logging.LoggerNameKey, name)
	return e
}

// WithValues satisfies our interface.
func (e *entryWrapper) WithValues(values map[string]interface{}) logging.Logger {
	e.entry = e.entry.WithFields(values)
	return e
}

// WithValue satisfies our interface.
func (e *entryWrapper) WithValue(key string, value interface{}) logging.Logger {
	e.entry = e.entry.WithField(key, value)
	return e
}

// WithError satisfies our interface.
func (e *entryWrapper) WithError(err error) logging.Logger {
	e.entry = e.entry.WithError(err)
	return e
}

// WithRequest satisfies our interface.
func (e *entryWrapper) WithRequest(req *http.Request) logging.Logger {
	return e.WithValues(map[string]interface{}{
		"path":   req.URL.Path,
		"method": req.Method,
		"query":  req.URL.RawQuery,
	})
}

// Info satisfies our interface.
func (e *entryWrapper) Info(input string) {
	e.entry.Infoln(input)
}

// Printf satisfies our interface.
func (e *entryWrapper) Printf(s string, args ...interface{}) {
	e.entry.Infoln(fmt.Sprintf(s, args...))
}

// Debug satisfies our interface.
func (e *entryWrapper) Debug(input string) {
	e.entry.Debugln(input)
}

// Error satisfies our interface.
func (e *entryWrapper) Error(err error, input string) {
	e.entry.Error(err, input)
}

// Fatal satisfies our interface.
func (e *entryWrapper) Fatal(err error) {
	e.entry.Fatal(err)
}
