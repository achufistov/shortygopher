package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

var gzipWriterPool = sync.Pool{
	New: func() interface{} {
		w, _ := gzip.NewWriterLevel(io.Discard, gzip.DefaultCompression)
		return w
	},
}

// gzipResponseWriter wraps http.ResponseWriter to provide gzip compression functionality.
// Implements transparent compression for supported content types.
type gzipResponseWriter struct {
	http.ResponseWriter
	gzWriter   *gzip.Writer
	shouldGzip bool
}

// WriteHeader writes the HTTP status code and sets up gzip compression if needed.
// Configures compression headers and initializes gzip writer from pool.
func (w *gzipResponseWriter) WriteHeader(statusCode int) {
	contentType := w.Header().Get("Content-Type")
	if w.shouldGzip && shouldCompress(contentType) {
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Add("Vary", "Accept-Encoding")

		w.gzWriter = gzipWriterPool.Get().(*gzip.Writer)
		w.gzWriter.Reset(w.ResponseWriter)
	}
	w.ResponseWriter.WriteHeader(statusCode)
}

// Write writes data to the response, compressing if gzip is enabled.
// Routes data through gzip writer if compression is active.
func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if w.gzWriter != nil {
		return w.gzWriter.Write(b)
	}
	return w.ResponseWriter.Write(b)
}

// Close closes the gzip writer and returns it to the pool for reuse.
// Must be called to properly clean up resources and avoid memory leaks.
func (w *gzipResponseWriter) Close() {
	if w.gzWriter != nil {
		w.gzWriter.Close()

		w.gzWriter.Reset(io.Discard)
		gzipWriterPool.Put(w.gzWriter)
		w.gzWriter = nil
	}
}

// shouldCompress determines if the given content type should be compressed.
// Returns true for text and JSON content types that benefit from compression.
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

// GzipMiddleware returns HTTP middleware that handles gzip compression for both requests and responses.
// Automatically decompresses incoming gzip requests and compresses outgoing responses when supported.
//
// Features:
//   - Decompresses incoming requests with Content-Encoding: gzip
//   - Compresses responses for clients that Accept-Encoding: gzip
//   - Uses sync.Pool for efficient gzip writer reuse
//   - Supports text/plain, application/json, and other compressible content types
//   - Handles application/x-gzip content type conversion
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
