package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Middleware для логирования
func LoggingMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Логируем информацию о запросе
			logger.Info("Request",
				zap.String("method", r.Method),
				zap.String("uri", r.RequestURI),
			)

			// Перехватываем ответ
			rw := &responseWriter{ResponseWriter: w}
			next.ServeHTTP(rw, r)

			// Логируем информацию о ответе
			logger.Info("Response",
				zap.Int("status", rw.status),
				zap.Int("size", rw.size),
				zap.Duration("duration", time.Since(start)),
			)
		})
	}
}

// responseWriter используется для перехвата статуса и размера ответа
type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.status = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}
