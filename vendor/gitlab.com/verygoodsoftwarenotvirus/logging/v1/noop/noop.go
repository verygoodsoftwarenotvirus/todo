package noop

import (
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v1"
)

// Logger is a default logger we can provide that does nothing in case of dire emergencies
type Logger struct{}

// ProvideNoopLogger provides our noop logger to dependency managers
func ProvideNoopLogger() logging.Logger {
	return &Logger{}
}

// Info fulfills our interface for the Info method
func (l *Logger) Info(string) {

}

// Debug fulfills our interface for the Debug method
func (l *Logger) Debug(string) {

}

// Error fulfills our interface for the Error method
func (l *Logger) Error(error, string) {

}

// Fatal fulfills our interface for the Fatal method
func (l *Logger) Fatal(error) {

}

// SetLevel fulfills our interface for the SetLevel method
func (l *Logger) SetLevel(logging.Level) {

}

// WithName fulfills our interface for the WithName method
func (l *Logger) WithName(string) logging.Logger {
	return l
}

// WithValues fulfills our interface for the WithValues method
func (l *Logger) WithValues(map[string]interface{}) logging.Logger {
	return l
}

// WithValue fulfills our interface for the WithValue method
func (l *Logger) WithValue(string, interface{}) logging.Logger {
	return l
}

// WithRequest fulfills our interface for the WithRequest method
func (l *Logger) WithRequest(*http.Request) logging.Logger {
	return l
}

// WithError fulfills our interface for the WithError method
func (l *Logger) WithError(error) logging.Logger {
	return l
}
