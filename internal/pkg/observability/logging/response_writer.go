package logging

import (
	"net/http"
	"time"
)

// ResponseWriter is a logging http.ResponseWriter.
type ResponseWriter struct {
	Wrapped http.ResponseWriter
	Logger  Logger

	writeCount uint
}

// Header mostly wraps the embedded ResponseWriter's Header.
func (rw *ResponseWriter) Header() http.Header {
	return rw.Wrapped.Header()
}

// Write mostly wraps the embedded ResponseWriter's Write.
func (rw *ResponseWriter) Write(b []byte) (int, error) {
	rw.writeCount += uint(len(b))
	return rw.Wrapped.Write(b)
}

// WriteHeader mostly wraps the embedded ResponseWriter's WriteHeader.
func (rw *ResponseWriter) WriteHeader(statusCode int) {
	rw.Logger = rw.Logger.WithValues(map[string]interface{}{
		"status_code":       statusCode,
		"header_written_on": time.Now().UnixNano(),
	})
	rw.Wrapped.WriteHeader(statusCode)
}
