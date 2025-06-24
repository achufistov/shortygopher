package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// LoggingMiddleware returns HTTP middleware that logs request and response information.
// Uses structured logging with zap to record HTTP method, URI, status, size, and duration.
//
// Logs two entries per request:
//   - Request: method and URI when request starts
//   - Response: status code, response size, and total duration
func LoggingMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			logger.Info("Request",
				zap.String("method", r.Method),
				zap.String("uri", r.RequestURI),
			)

			rw := &responseWriter{ResponseWriter: w}
			next.ServeHTTP(rw, r)

			logger.Info("Response",
				zap.Int("status", rw.status),
				zap.Int("size", rw.size),
				zap.Duration("duration", time.Since(start)),
			)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture response status and size.
// Used by logging middleware to record response metadata.
type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

// WriteHeader captures the HTTP status code for logging.
func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.status = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// Write captures the response size and forwards data to the underlying writer.
func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}
