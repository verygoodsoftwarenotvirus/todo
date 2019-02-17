package logging

import (
	"net/http"
	"time"
)

// ResponseWriter is a logging http.ResponseWriter
type ResponseWriter struct {
	Wrapped http.ResponseWriter
	Logger  Logger

	writeCount uint
}

// Header mostly wraps the embedded ResponseWriter's Header
func (rw *ResponseWriter) Header() http.Header {
	return rw.Wrapped.Header()
}

// Write mostly wraps the embedded ResponseWriter's Write
func (rw *ResponseWriter) Write(b []byte) (int, error) {
	rw.writeCount += uint(len(b))
	return rw.Wrapped.Write(b)
}

// WriteHeader mostly wraps the embedded ResponseWriter's WriteHeader
func (rw *ResponseWriter) WriteHeader(statusCode int) {
	rw.Logger = rw.Logger.WithValues(map[string]interface{}{
		"status_code":       statusCode,
		"header_written_on": time.Now().UnixNano(),
	})
	rw.Wrapped.WriteHeader(statusCode)
}

// BuildMiddleware builds a logging middleware
func BuildMiddleware(logger Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			lrw := &ResponseWriter{
				Wrapped: w,
				Logger: logger.WithValues(map[string]interface{}{
					"request_id": r.Header.Get("X-Request-Id"),
					"method":     r.Method,
					"url":        r.URL.String(),
					"host_ip":    r.Host,
				}),
			}
			w = lrw

			defer func() {
				values := map[string]interface{}{
					"latency":      time.Since(start).String(),
					"content_size": lrw.writeCount,
				}

				lrw.Logger.WithValues(values).Debug("")
			}()

			next.ServeHTTP(w, r)
		})
	}
}
