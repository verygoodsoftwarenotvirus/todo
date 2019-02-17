package logging

import (
	"net/http"
)

// Level is a simple string alias
type Level string

var (
	// InfoLevel describes a info-level log
	InfoLevel Level = "info"
	// DebugLevel describes a debug-level log
	DebugLevel Level = "debug"
	// ErrorLevel describes a error-level log
	ErrorLevel Level = "error"
)

// Logger represents a simple logging interface we can build wrappers around.
type Logger interface {
	Info(string)
	Debug(string)
	Error(error, string)
	Fatal(error)
	Print(...interface{})

	SetLevel(Level)

	// Builder funcs

	WithValues(map[string]interface{}) Logger
	WithValue(string, interface{}) Logger
	WithRequest(*http.Request) Logger
	WithError(error) Logger
}

//
//// LogFormatter formats logs for our chosen router, chi
//type LogFormatter struct {
//	logger zerolog.Logger
//}
//
//// ProvideLogFormatter provides a new log formatter
//func ProvideLogFormatter() *LogFormatter {
//	w := diode.NewWriter(os.Stdout, 1000, 10*time.Millisecond, func(missed int) {
//		log.Printf("logger dropped %d messages\n", missed)
//	})
//	logger := zerolog.New(w).With().Caller().Timestamp().Logger()
//	lf := &LogFormatter{
//		logger: logger,
//	}
//	return lf
//}
//
//// LogEntry represents a log entry
//type LogEntry struct {
//	requestID string
//	logger    zerolog.Logger
//}
//
//// Write helps fulfill our middleware.LogEntry interface
//func (e LogEntry) Write(status, bytes int, elapsed time.Duration) {
//	logger := e.logger.
//		With().
//		Str("request_id", e.requestID).
//		Int("status", status).
//		Dur("elapsed", elapsed).
//		Int("wrote", bytes).
//		Logger()
//	logger.Debug().Msg("")
//}
//
//// Panic helps fulfill our middleware.LogEntry interface
//func (e LogEntry) Panic(v interface{}, stack []byte) {
//	logger := e.logger.With().Interface("v", v).Bytes("stack", stack).Logger()
//	logger.Panic()
//}
//
//// NewLogEntry creates a new LogEntry for the request.
//func (f *LogFormatter) NewLogEntry(r *http.Request) middleware.LogEntry {
//	entry := &LogEntry{
//		logger:    f.logger,
//		requestID: middleware.GetReqID(r.Context()),
//	}
//
//	return entry
//}
