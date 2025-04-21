package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"
)

type gzipResponseWriter struct {
	http.ResponseWriter
	gzWriter   *gzip.Writer
	shouldGzip bool
}

func (w *gzipResponseWriter) WriteHeader(statusCode int) {
	contentType := w.Header().Get("Content-Type")
	if w.shouldGzip && shouldCompress(contentType) {
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Add("Vary", "Accept-Encoding")
		w.gzWriter = gzip.NewWriter(w.ResponseWriter)
	}
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if w.gzWriter != nil {
		return w.gzWriter.Write(b)
	}
	return w.ResponseWriter.Write(b)
}

func (w *gzipResponseWriter) Close() {
	if w.gzWriter != nil {
		w.gzWriter.Close()
	}
}

func shouldCompress(contentType string) bool {
	compressibleTypes := []string{
		"text/plain",
		"text/html",
		"application/json",
		"application/javascript",
		"text/css",
		"application/xml",
	}

	for _, t := range compressibleTypes {
		if strings.HasPrefix(contentType, t) {
			return true
		}
	}
	return false
}

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// process incoming gzip
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gzReader, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Invalid gzip body", http.StatusBadRequest)
				return
			}
			defer gzReader.Close()
			r.Body = gzReader
		}

		if r.Header.Get("Content-Type") == "application/x-gzip" {
			r.Header.Set("Content-Type", "text/plain")
		}

		acceptsGzip := strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")

		gzw := &gzipResponseWriter{
			ResponseWriter: w,
			shouldGzip:     acceptsGzip,
		}
		defer gzw.Close()

		next.ServeHTTP(gzw, r)
	})
}
