package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// GzipResponseWriter оборачивает http.ResponseWriter для сжатия ответа
type GzipResponseWriter struct {
	http.ResponseWriter
	Writer *gzip.Writer
}

func (w *GzipResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *GzipResponseWriter) Write(b []byte) (int, error) {
	if w.Writer != nil {
		return w.Writer.Write(b)
	}
	// If Writer is nil, just write to the ResponseWriter directly
	return w.ResponseWriter.Write(b)
}

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle incoming gzip-encoded requests
		if r.Header.Get("Content-Encoding") == "gzip" {
			reader, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Failed to create gzip reader", http.StatusBadRequest)
				return
			}
			defer reader.Close()

			body, err := io.ReadAll(reader)
			if err != nil {
				http.Error(w, "Failed to read gzip body", http.StatusBadRequest)
				return
			}

			// Replace the request body with the decompressed data
			r.Body = io.NopCloser(bytes.NewBuffer(body))
			r.Header.Set("Content-Length", strconv.Itoa(len(body)))

			// If the Content-Type is application/x-gzip, assume the body is plain text
			if r.Header.Get("Content-Type") == "application/x-gzip" {
				r.Header.Set("Content-Type", "text/plain")
			}
		}

		// Handle outgoing gzip-encoded responses
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Set("Vary", "Accept-Encoding")

			gzipWriter := gzip.NewWriter(w)
			defer gzipWriter.Close()

			next.ServeHTTP(&GzipResponseWriter{ResponseWriter: w, Writer: gzipWriter}, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}
