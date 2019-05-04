package logrus

import (
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/logging/v1"

	"github.com/sirupsen/logrus"
)

// entryWrapper has repeats of many functions for the sake of interface implementation
type entryWrapper struct {
	*logrus.Entry
}

// SetLevel is a noop used for the sake of interface implementation
// if you're calling this on an entrywrapper, then the level has
// already been determined. I have no use for a function which
// overrides that behavior, simple though it may be to paste it here
func (e *entryWrapper) SetLevel(level logging.Level) {}

// WithName is our obligatory contract fulfillment function
// Logrus doesn't support named loggers :(
func (e *entryWrapper) WithName(name string) logging.Logger {
	return &entryWrapper{e.WithField(logging.LoggerNameKey, name)}
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
func (e *entryWrapper) Error(err error, input string) {
	e.Error(err, input)
}

// Fatal satisfies our contract for the logging.Logger Fatal method.
func (e *entryWrapper) Fatal(err error) {
	e.Fatal(err)
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

// WithRequest satisfies our contract for the logging.Logger WithRequest method.
func (e *entryWrapper) WithRequest(req *http.Request) logging.Logger {
	return e.WithValues(map[string]interface{}{
		"path":   req.URL.Path,
		"method": req.Method,
		"query":  req.URL.RawQuery,
	})
}
