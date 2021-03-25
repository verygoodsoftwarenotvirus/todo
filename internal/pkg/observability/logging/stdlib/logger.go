package stdlib

import (
	"io"
	"log"
	"net/http"
	"os"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
)

// logger is our log wrapper.
type logger struct {
	requestIDFunc logging.RequestIDFunc
	logger        *log.Logger
}

func buildNewStdLogger(out io.Writer, prefix string) *log.Logger {
	return log.New(out, prefix, log.LstdFlags)
}

// NewLogger builds a new logger.
func NewLogger(out io.Writer, prefix string) logging.Logger {
	return &logger{logger: buildNewStdLogger(out, prefix)}
}

// NewLoggerWithSource builds a new logger.
func NewLoggerWithSource(l *log.Logger) logging.Logger {
	if l == nil {
		l = buildNewStdLogger(os.Stderr, "")
	}

	return &logger{logger: l}
}

// WithName is our obligatory contract fulfillment function.
// Zerolog doesn't support named loggers :( so we have this workaround.
func (l *logger) WithName(name string) logging.Logger {
	l.logger.SetPrefix(name)
	return l
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
	l.logger.Println(input)
}

// Debug satisfies our contract for the logging.Logger Debug method.
func (l *logger) Debug(input string) {
	l.logger.Println(input)
}

// Error satisfies our contract for the logging.Logger Error method.
func (l *logger) Error(err error, input string) {
	l.logger.Println(err, input)
}

// Fatal satisfies our contract for the logging.Logger Fatal method.
func (l *logger) Fatal(err error) {
	l.logger.Fatal(err)
}

// Printf satisfies our contract for the logging.Logger Printf method.
func (l *logger) Printf(format string, args ...interface{}) {
	l.logger.Printf(format, args...)
}

// Clone satisfies our contract for the logging.Logger WithValues method.
func (l *logger) Clone() logging.Logger {
	return NewLoggerWithSource(l.logger)
}

// WithValues satisfies our contract for the logging.Logger WithValues method.
func (l *logger) WithValues(map[string]interface{}) logging.Logger {
	return l
}

// WithValue satisfies our contract for the logging.Logger WithValue method.
func (l *logger) WithValue(string, interface{}) logging.Logger {
	return l
}

// WithError satisfies our contract for the logging.Logger WithError method.
func (l *logger) WithError(error) logging.Logger {
	return l
}

// WithRequest satisfies our contract for the logging.Logger WithRequest method.
func (l *logger) WithRequest(*http.Request) logging.Logger {
	return l
}

// WithResponse satisfies our contract for the logging.Logger WithRequest method.
func (l *logger) WithResponse(*http.Response) logging.Logger {
	return l
}
