package logrus

import (
	"net/http"

	"github.com/sirupsen/logrus"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
)

var _ logging.Logger = (*logger)(nil)

// logger is our log wrapper.
type logger struct {
	logger        *logrus.Logger
	requestIDFunc logging.RequestIDFunc
}

// buildLogrus builds a logrus logger to our specs.
func buildLogrus() *logrus.Logger {
	l := logrus.New()

	l.SetFormatter(&logrus.JSONFormatter{
		DataKey: "meta",
		// PrettyPrint: true,
	})

	l.SetReportCaller(true)

	return l
}

// NewLogger builds a new logrus-backed logger.
func NewLogger(l *logrus.Logger) logging.Logger {
	if l == nil {
		l = buildLogrus()
	}

	return &logger{logger: l}
}

// WithName is our obligatory contract fulfillment function
// Logrus doesn't support named loggers.
func (l *logger) WithName(name string) logging.Logger {
	return &entryWrapper{entry: l.logger.WithField(logging.LoggerNameKey, name)}
}

// SetLevel sets the log level.
func (l *logger) SetLevel(level logging.Level) {
	var lvl logrus.Level

	switch level {
	case logging.InfoLevel:
		lvl = logrus.InfoLevel
	case logging.DebugLevel:
		lvl = logrus.DebugLevel
	case logging.WarnLevel:
		lvl = logrus.WarnLevel
	case logging.ErrorLevel:
		lvl = logrus.ErrorLevel
	}

	l.logger.SetLevel(lvl)
}

// SetRequestIDFunc satisfies our interface.
func (l *logger) SetRequestIDFunc(f logging.RequestIDFunc) {
	l.requestIDFunc = f
}

// Info satisfies our contract for the logging.Logger Info method.
func (l *logger) Info(input string) {
	l.logger.Infoln(input)
}

// Debug satisfies our contract for the logging.Logger Debug method.
func (l *logger) Debug(input string) {
	l.logger.Debugln(input)
}

// Error satisfies our contract for the logging.Logger Error method.
func (l *logger) Error(err error, input string) {
	l.logger.WithField("err", err).Errorln(input)
}

// Fatal satisfies our contract for the logging.Logger Fatal method.
func (l *logger) Fatal(err error) {
	l.logger.WithField("err", err).Fatal()
}

// Printf satisfies our contract for the logging.Logger Printf method.
func (l *logger) Printf(format string, args ...interface{}) {
	l.logger.Printf(format, args...)
}

// WithValues satisfies our contract for the logging.Logger WithValues method.
func (l *logger) WithValues(values map[string]interface{}) logging.Logger {
	return &entryWrapper{entry: l.logger.WithFields(values)}
}

// WithValue satisfies our contract for the logging.Logger WithValue method.
func (l *logger) WithValue(key string, value interface{}) logging.Logger {
	return &entryWrapper{entry: l.logger.WithField(key, value)}
}

// WithError satisfies our contract for the logging.Logger WithError method.
func (l *logger) WithError(err error) logging.Logger {
	return &entryWrapper{entry: l.logger.WithError(err)}
}

func (l *logger) attachRequestToLog(req *http.Request) *logrus.Logger {
	if req != nil {
		l2 := l.logger.WithFields(map[string]interface{}{
			"path":   req.URL.Path,
			"method": req.Method,
			"query":  req.URL.RawQuery,
		})

		if l.requestIDFunc != nil {
			if reqID := l.requestIDFunc(req); reqID != "" {
				l2 = l2.WithField("request_id", reqID)
			}
		}

		return l2.Logger
	}

	return l.logger
}

// WithRequest satisfies our contract for the logging.Logger WithRequest method.
func (l *logger) WithRequest(req *http.Request) logging.Logger {
	return &logger{logger: l.attachRequestToLog(req)}
}

// WithResponse satisfies our contract for the logging.Logger WithResponse method.
func (l *logger) WithResponse(res *http.Response) logging.Logger {
	l2 := l.attachRequestToLog(res.Request).WithField(keys.ResponseStatusKey, res.StatusCode).Logger

	return &logger{logger: l2}
}
