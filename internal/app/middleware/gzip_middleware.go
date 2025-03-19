package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// GzipMiddleware сжимает ответы и распаковывает запросы, если это необходимо.
func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, поддерживает ли клиент gzip
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			// Создаем gzip.Writer поверх текущего http.ResponseWriter
			gz := gzip.NewWriter(w)
			defer gz.Close()

			w.Header().Set("Content-Encoding", "gzip")
			w = &gzipResponseWriter{ResponseWriter: w, Writer: gz}
		}

		// Проверяем, сжат ли запрос
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			// Если Content-Type не соответствует ожидаемому, просто продолжим
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Failed to read gzip body", http.StatusBadRequest)
				return
			}
			defer gz.Close()
			r.Body = gz
		}

		next.ServeHTTP(w, r)
	})
}

// gzipResponseWriter оборачивает http.ResponseWriter для сжатия данных
type gzipResponseWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}
