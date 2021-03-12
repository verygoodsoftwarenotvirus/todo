package stdlib

import (
	"io"
	"log"
	"net/http"
	"os"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
)

// Logger is our log wrapper.
type Logger struct {
	requestIDFunc logging.RequestIDFunc
	logger        *log.Logger
}

func buildNewStdLogger(out io.Writer, prefix string) *log.Logger {
	return log.New(out, prefix, log.LstdFlags)
}

// NewLogger builds a new logger.
func NewLogger(out io.Writer, prefix string) logging.Logger {
	return &Logger{logger: buildNewStdLogger(out, prefix)}
}

// NewLoggerWithSource builds a new logger.
func NewLoggerWithSource(logger *log.Logger) logging.Logger {
	if logger == nil {
		logger = buildNewStdLogger(os.Stderr, "")
	}

	return &Logger{logger: logger}
}

// WithName is our obligatory contract fulfillment function.
// Zerolog doesn't support named loggers :( so we have this workaround.
func (l *Logger) WithName(name string) logging.Logger {
	l.logger.SetPrefix(name)
	return l
}

// SetLevel sets the log level for our logger.
func (l *Logger) SetLevel(_ logging.Level) {}

// SetRequestIDFunc sets the request ID retrieval function.
func (l *Logger) SetRequestIDFunc(f logging.RequestIDFunc) {
	if f != nil {
		l.requestIDFunc = f
	}
}

// Info satisfies our contract for the logging.Logger Info method.
func (l *Logger) Info(input string) {
	l.logger.Println(input)
}

// Debug satisfies our contract for the logging.Logger Debug method.
func (l *Logger) Debug(input string) {
	l.logger.Println(input)
}

// Error satisfies our contract for the logging.Logger Error method.
func (l *Logger) Error(err error, input string) {
	l.logger.Println(err, input)
}

// Fatal satisfies our contract for the logging.Logger Fatal method.
func (l *Logger) Fatal(err error) {
	l.logger.Fatal(err)
}

// Printf satisfies our contract for the logging.Logger Printf method.
func (l *Logger) Printf(format string, args ...interface{}) {
	l.logger.Printf(format, args...)
}

// WithValues satisfies our contract for the logging.Logger WithValues method.
func (l *Logger) WithValues(_ map[string]interface{}) logging.Logger {
	return l
}

// WithValue satisfies our contract for the logging.Logger WithValue method.
func (l *Logger) WithValue(_ string, _ interface{}) logging.Logger {
	return l
}

// WithError satisfies our contract for the logging.Logger WithError method.
func (l *Logger) WithError(_ error) logging.Logger {
	return l
}

// WithRequest satisfies our contract for the logging.Logger WithRequest method.
func (l *Logger) WithRequest(_ *http.Request) logging.Logger {
	return l
}
