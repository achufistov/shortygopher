package shortygopher_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	"github.com/achufistov/shortygopher.git/internal/app/config"
	"github.com/achufistov/shortygopher.git/internal/app/handlers"
	"github.com/achufistov/shortygopher.git/internal/app/middleware"
	"github.com/achufistov/shortygopher.git/internal/app/service"
	"github.com/achufistov/shortygopher.git/internal/app/storage"
	"github.com/go-chi/chi/v5"
)

// ExampleHandlePost demonstrates basic usage of the handler for URL shortening
// in text format via POST /.
func ExampleHandlePost() {
	// Create a temporary JWT secret for the example
	tmpfile, err := os.CreateTemp("", "example_secret")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte("example-secret-key")); err != nil {
		log.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		log.Fatal(err)
	}

	// Set environment variable for JWT secret
	os.Setenv("JWT_SECRET_FILE", tmpfile.Name())
	defer os.Unsetenv("JWT_SECRET_FILE")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Create storage and initialize handlers
	storageInstance := storage.NewURLStorage()
	handlers.InitStorage(storageInstance)
	
	// Initialize service
	serviceInstance := service.NewService(storageInstance, cfg)
	handlers.InitService(serviceInstance)

	// Create HTTP request
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com"))
	req.Header.Set("Content-Type", "text/plain")

	// Add context with user
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "example-user")
	req = req.WithContext(ctx)

	// Create ResponseWriter
	w := httptest.NewRecorder()

	// Call handler
	handlers.HandlePost(cfg, w, req)

	// Output result
	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Content-Type: %s\n", w.Header().Get("Content-Type"))
	fmt.Printf("URL shortened: %t\n", strings.Contains(w.Body.String(), cfg.BaseURL))

	// Output:
	// Status: 201
	// Content-Type: text/plain
	// URL shortened: true
}

// ExampleHandleShortenPost demonstrates using JSON API
// for URL shortening via POST /api/shorten.
func ExampleHandleShortenPost() {
	// Create a temporary JWT secret for the example
	tmpfile, err := os.CreateTemp("", "example_secret")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte("example-secret-key")); err != nil {
		log.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		log.Fatal(err)
	}

	os.Setenv("JWT_SECRET_FILE", tmpfile.Name())
	defer os.Unsetenv("JWT_SECRET_FILE")

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	storageInstance := storage.NewURLStorage()
	handlers.InitStorage(storageInstance)
	
	// Initialize service
	serviceInstance := service.NewService(storageInstance, cfg)
	handlers.InitService(serviceInstance)

	// Create JSON request
	reqBody := handlers.ShortenRequest{
		OriginalURL: "https://example.com/very/long/path",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "example-user")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handlers.HandleShortenPost(cfg, w, req)

	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Content-Type: %s\n", w.Header().Get("Content-Type"))

	var resp handlers.ShortenResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	fmt.Printf("Shortened URL received: %t\n", strings.Contains(resp.ShortURL, cfg.BaseURL))

	// Output:
	// Status: 201
	// Content-Type: application/json
	// Shortened URL received: true
}

// ExampleHandleBatchShortenPost demonstrates batch shortening multiple URLs
// via POST /api/shorten/batch.
func ExampleHandleBatchShortenPost() {
	tmpfile, err := os.CreateTemp("", "example_secret")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte("example-secret-key")); err != nil {
		log.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		log.Fatal(err)
	}

	os.Setenv("JWT_SECRET_FILE", tmpfile.Name())
	defer os.Unsetenv("JWT_SECRET_FILE")

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	storageInstance := storage.NewURLStorage()
	handlers.InitStorage(storageInstance)
	
	// Initialize service
	serviceInstance := service.NewService(storageInstance, cfg)
	handlers.InitService(serviceInstance)

	// Create batch request
	batchReq := []handlers.BatchRequest{
		{CorrelationID: "1", OriginalURL: "https://example.com"},
		{CorrelationID: "2", OriginalURL: "https://google.com"},
	}
	jsonBody, _ := json.Marshal(batchReq)

	req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "example-user")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handlers.HandleBatchShortenPost(cfg, w, req)

	fmt.Printf("Status: %d\n", w.Code)

	var resp []handlers.BatchResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	fmt.Printf("URLs processed: %d\n", len(resp))
	fmt.Printf("First correlation ID: %s\n", resp[0].CorrelationID)

	// Output:
	// Status: 201
	// URLs processed: 2
	// First correlation ID: 1
}

// ExampleHandleGet demonstrates getting the original URL
// and redirect via GET /{id}.
func ExampleHandleGet() {
	cfg := &config.Config{
		Address:     ":8080",
		BaseURL:     "http://localhost:8080",
		FileStorage: "",
	}
	
	storageInstance := storage.NewURLStorage()
	handlers.InitStorage(storageInstance)
	
	// Initialize service
	serviceInstance := service.NewService(storageInstance, cfg)
	handlers.InitService(serviceInstance)

	// First, add URL to storage
	shortURL := "example123"
	originalURL := "https://example.com"
	storageInstance.AddURL(shortURL, originalURL, "example-user")

	// Create GET request
	req := httptest.NewRequest(http.MethodGet, "/"+shortURL, nil)

	// Add route parameter
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", shortURL)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handlers.HandleGet(w, req)

	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Location: %s\n", w.Header().Get("Location"))

	// Output:
	// Status: 307
	// Location: https://example.com
}
