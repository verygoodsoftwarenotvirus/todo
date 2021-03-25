package testing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
)

// logger is our log wrapper.
type logger struct {
	requestIDFunc logging.RequestIDFunc
	contextValues map[string]interface{}
	t             *testing.T
	contextValHat sync.RWMutex
}

// newLogger builds a new logger.
func newLogger(t *testing.T, ctxVals map[string]interface{}) *logger {
	v := map[string]interface{}{}
	if ctxVals != nil {
		v = ctxVals
	}

	return &logger{
		t:             t,
		contextValues: v,
	}
}

// NewLogger builds a new logger.
func NewLogger(t *testing.T) logging.Logger {
	return newLogger(t, nil)
}

// clone clones the current logger.
func (l *logger) clone() *logger {
	return &logger{
		t:             l.t,
		contextValues: l.contextValues,
	}
}

func (l *logger) jsonString(msg string, args ...interface{}) string {
	l.contextValHat.Lock()
	defer l.contextValHat.Unlock()

	m := l.contextValues
	m["msg"] = fmt.Sprintf(msg, args...)

	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(m); err != nil {
		panic(err)
	}

	return b.String()
}

// WithName is our obligatory contract fulfillment function.
func (l *logger) WithName(name string) logging.Logger {
	return l.WithValue(logging.LoggerNameKey, name)
}

// SetLevel sets the log level for our logger.
func (l *logger) SetLevel(_ logging.Level) {}

// SetRequestIDFunc sets the request ID retrieval function.
func (l *logger) SetRequestIDFunc(f logging.RequestIDFunc) {
	if f != nil {
		l.requestIDFunc = f
	}
}

// Info satisfies our contract for the logging.Logger Info method.
func (l *logger) Info(input string) {
	l.t.Helper()

	l.t.Logf("%s", l.jsonString(input))
}

// Debug satisfies our contract for the logging.Logger Debug method.
func (l *logger) Debug(input string) {
	l.t.Helper()

	l.t.Logf("%s", l.jsonString(input))
}

// Error satisfies our contract for the logging.Logger Error method.
func (l *logger) Error(err error, input string) {
	l.t.Helper()

	l.withValue("error", err).t.Logf(l.jsonString(input))
}

// Fatal satisfies our contract for the logging.Logger Fatal method.
func (l *logger) Fatal(err error) {
	l.t.Helper()

	l.t.Logf("%s", l.jsonString(err.Error()))
}

// Printf satisfies our contract for the logging.Logger Printf method.
func (l *logger) Printf(format string, args ...interface{}) {
	l.t.Helper()

	l.t.Logf("%s", l.jsonString(format, args...))
}

// Clone satisfies our contract for the logging.Logger WithValue method.
func (l *logger) Clone() logging.Logger {
	return &logger{t: l.t}
}

// WithValues satisfies our contract for the logging.Logger WithValues method.
func (l *logger) WithValues(values map[string]interface{}) logging.Logger {
	l2 := l.clone()

	l2.contextValHat.Lock()
	defer l2.contextValHat.Unlock()

	for k, v := range values {
		l2.contextValues[k] = v
	}

	return l2
}

func (l *logger) withValue(k string, v interface{}) *logger {
	l2 := l.clone()

	switch x := v.(type) {
	case string:
		withoutTabs := strings.ReplaceAll(x, "\t", "\\t")
		withoutNewlines := strings.ReplaceAll(withoutTabs, "\n", "\\n")

		v = withoutNewlines
	}

	l2.contextValHat.Lock()
	defer l2.contextValHat.Unlock()
	l2.contextValues[k] = v

	return l2
}

// WithValue satisfies our contract for the logging.Logger WithValue method.
func (l *logger) WithValue(k string, v interface{}) logging.Logger {
	return l.withValue(k, v)
}

// WithError satisfies our contract for the logging.Logger WithError method.
func (l *logger) WithError(err error) logging.Logger {
	return l.withValue("error", err)
}

// WithRequest satisfies our contract for the logging.Logger WithRequest method.
func (l *logger) WithRequest(req *http.Request) logging.Logger {
	if req != nil {
		return l.withValue("path", req.URL.Path).
			withValue("method", req.Method).
			withValue("query", req.URL.RawQuery)
	}

	return l
}

// WithResponse satisfies our contract for the logging.Logger WithRequest method.
func (l *logger) WithResponse(res *http.Response) logging.Logger {
	if res != nil {
		return l.withValue("status_code", res.StatusCode).WithRequest(res.Request)
	}

	return l
}
