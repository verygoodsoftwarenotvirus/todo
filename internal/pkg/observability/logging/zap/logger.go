package zap

import (
	"fmt"
	"net/http"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is our log wrapper.
type Logger struct {
	logger *zap.Logger
}

// NewLogger builds a new logger.
func NewLogger(logger *zap.Logger) (logging.Logger, error) {
	var (
		l   *zap.Logger
		err error
	)

	if logger == nil {
		l, err = zap.NewDevelopment()
		if err != nil {
			return nil, err
		}
	} else {
		l = logger
	}

	return &Logger{logger: l}, nil
}

// WithName is our obligatory contract fulfillment function.
func (l *Logger) WithName(name string) logging.Logger {
	l2 := l.logger.Named(name)
	return &Logger{logger: l2}
}

// SetLevel sets the log level for our logger.
func (l *Logger) SetLevel(level logging.Level) {
	var lvl zapcore.Level

	switch level {
	case logging.InfoLevel:
		lvl = zapcore.InfoLevel
	case logging.DebugLevel:
		lvl = zapcore.DebugLevel
	case logging.WarnLevel:
		lvl = zapcore.WarnLevel
	case logging.ErrorLevel:
		lvl = zapcore.ErrorLevel
	}

	l.logger = l.logger.WithOptions(zap.IncreaseLevel(lvl))
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
	l.logger.With(zap.Error(err)).Error(input)
}

// Fatal satisfies our contract for the logging.Logger Fatal method.
func (l *Logger) Fatal(err error) {
	l.logger.With(zap.Error(err)).Fatal(err.Error())
}

// Printf satisfies our contract for the logging.Logger Printf method.
func (l *Logger) Printf(format string, args ...interface{}) {
	l.logger.Info(fmt.Sprintf(format, args...))
}

func determineField(key string, val interface{}) zap.Field {
	switch x := val.(type) {
	case bool:
		return zap.Bool(key, x)
	case *bool:
		return zap.Boolp(key, x)
	case []bool:
		return zap.Bools(key, x)
	case []byte:
		return zap.ByteString(key, x)
	case [][]byte:
		return zap.ByteStrings(key, x)
	case complex128:
		return zap.Complex128(key, x)
	case *complex128:
		return zap.Complex128p(key, x)
	case []complex128:
		return zap.Complex128s(key, x)
	case complex64:
		return zap.Complex64(key, x)
	case *complex64:
		return zap.Complex64p(key, x)
	case []complex64:
		return zap.Complex64s(key, x)
	case time.Duration:
		return zap.Duration(key, x)
	case *time.Duration:
		return zap.Durationp(key, x)
	case []time.Duration:
		return zap.Durations(key, x)
	case float32:
		return zap.Float32(key, x)
	case *float32:
		return zap.Float32p(key, x)
	case []float32:
		return zap.Float32s(key, x)
	case float64:
		return zap.Float64(key, x)
	case *float64:
		return zap.Float64p(key, x)
	case []float64:
		return zap.Float64s(key, x)
	case int:
		return zap.Int(key, x)
	case int16:
		return zap.Int16(key, x)
	case *int16:
		return zap.Int16p(key, x)
	case []int16:
		return zap.Int16s(key, x)
	case int32:
		return zap.Int32(key, x)
	case *int32:
		return zap.Int32p(key, x)
	case []int32:
		return zap.Int32s(key, x)
	case int64:
		return zap.Int64(key, x)
	case *int64:
		return zap.Int64p(key, x)
	case []int64:
		return zap.Int64s(key, x)
	case int8:
		return zap.Int8(key, x)
	case *int8:
		return zap.Int8p(key, x)
	case []int8:
		return zap.Int8s(key, x)
	case *int:
		return zap.Intp(key, x)
	case []int:
		return zap.Ints(key, x)
	case time.Time:
		return zap.Time(key, x)
	case *time.Time:
		return zap.Timep(key, x)
	case []time.Time:
		return zap.Times(key, x)
	case uint:
		return zap.Uint(key, x)
	case uint16:
		return zap.Uint16(key, x)
	case *uint16:
		return zap.Uint16p(key, x)
	case []uint16:
		return zap.Uint16s(key, x)
	case uint32:
		return zap.Uint32(key, x)
	case *uint32:
		return zap.Uint32p(key, x)
	case []uint32:
		return zap.Uint32s(key, x)
	case uint64:
		return zap.Uint64(key, x)
	case *uint64:
		return zap.Uint64p(key, x)
	case []uint64:
		return zap.Uint64s(key, x)
	case uint8:
		return zap.Uint8(key, x)
	case *uint8:
		return zap.Uint8p(key, x)
	case *uint:
		return zap.Uintp(key, x)
	case uintptr:
		return zap.Uintptr(key, x)
	case *uintptr:
		return zap.Uintptrp(key, x)
	case []uintptr:
		return zap.Uintptrs(key, x)
	case []uint:
		return zap.Uints(key, x)
	case string:
		return zap.String(key, x)
	case *string:
		return zap.Stringp(key, x)
	case []string:
		return zap.Strings(key, x)
	case fmt.Stringer:
		return zap.Stringer(key, x)
	default:
		return zap.Any(key, val)
	}
}

// WithValues satisfies our contract for the logging.Logger WithValues method.
func (l *Logger) WithValues(values map[string]interface{}) logging.Logger {
	l2 := l.logger.With()

	for key, val := range values {
		l2 = l2.With(determineField(key, val))
	}

	return &Logger{logger: l2}
}

// WithValue satisfies our contract for the logging.Logger WithValue method.
func (l *Logger) WithValue(key string, value interface{}) logging.Logger {
	l2 := l.logger.With(determineField(key, value))
	return &Logger{logger: l2}
}

// WithError satisfies our contract for the logging.Logger WithError method.
func (l *Logger) WithError(err error) logging.Logger {
	l2 := l.logger.With(zap.Error(err))
	return &Logger{logger: l2}
}

// WithRequest satisfies our contract for the logging.Logger WithRequest method.
func (l *Logger) WithRequest(req *http.Request) logging.Logger {
	return &Logger{logger: l.logger.With(
		zap.String("path", req.URL.Path),
		zap.String("method", req.Method),
		zap.String("query", req.URL.RawQuery),
	)}
}
