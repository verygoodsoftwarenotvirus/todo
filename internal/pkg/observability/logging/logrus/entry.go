package logrus

import (
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
)

// entryWrapper has repeats of many functions for the sake of interface implementation.
type entryWrapper struct {
	entry         *logrus.Entry
	requestIDFunc logging.RequestIDFunc
}

// SetLevel is a noop used for the sake of interface implementation
// if you're calling this on an entry wrapper, then the level has
// already been determined. I have no use for a function which
// overrides that behavior, simple though it may be to paste it here.
func (w *entryWrapper) SetLevel(level logging.Level) {
	switch level {
	case logging.InfoLevel:
		w.entry.Level = logrus.InfoLevel
	case logging.DebugLevel:
		w.entry.Level = logrus.DebugLevel
	case logging.ErrorLevel:
		w.entry.Level = logrus.ErrorLevel
	case logging.WarnLevel:
		w.entry.Level = logrus.WarnLevel
	}
}

// SetRequestIDFunc satisfies our interface.
func (w *entryWrapper) SetRequestIDFunc(f logging.RequestIDFunc) {
	w.requestIDFunc = f
}

// WithName is our obligatory contract fulfillment function. Logrus doesn't support named loggers.
func (w *entryWrapper) WithName(name string) logging.Logger {
	w.entry = w.entry.WithField(logging.LoggerNameKey, name)
	return w
}

// Clone satisfies our interface.
func (w *entryWrapper) Clone() logging.Logger {
	return &entryWrapper{
		entry:         w.entry,
		requestIDFunc: w.requestIDFunc,
	}
}

// WithValues satisfies our interface.
func (w *entryWrapper) WithValues(values map[string]interface{}) logging.Logger {
	w.entry = w.entry.WithFields(values)
	return w
}

// WithValue satisfies our interface.
func (w *entryWrapper) WithValue(key string, value interface{}) logging.Logger {
	w.entry = w.entry.WithField(key, value)
	return w
}

// WithError satisfies our interface.
func (w *entryWrapper) WithError(err error) logging.Logger {
	w.entry = w.entry.WithError(err)
	return w
}

// WithRequest satisfies our interface.
func (w *entryWrapper) WithRequest(req *http.Request) logging.Logger {
	if req != nil {
		return w.WithValues(map[string]interface{}{
			"path":   req.URL.Path,
			"method": req.Method,
			"query":  req.URL.RawQuery,
		})
	}

	return w
}

// WithResponse satisfies our interface.
func (w *entryWrapper) WithResponse(res *http.Response) logging.Logger {
	w2 := w.WithRequest(res.Request)

	w2 = w2.WithValue(keys.ResponseStatusKey, res.StatusCode)

	return w2
}

// Info satisfies our interface.
func (w *entryWrapper) Info(input string) {
	w.entry.Infoln(input)
}

// Printf satisfies our interface.
func (w *entryWrapper) Printf(s string, args ...interface{}) {
	w.entry.Infoln(fmt.Sprintf(s, args...))
}

// Debug satisfies our interface.
func (w *entryWrapper) Debug(input string) {
	w.entry.Debugln(input)
}

// Error satisfies our interface.
func (w *entryWrapper) Error(err error, input string) {
	w.entry.Error(err, input)
}

// Fatal satisfies our interface.
func (w *entryWrapper) Fatal(err error) {
	w.entry.Fatal(err)
}
