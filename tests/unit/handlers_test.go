package unit_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/achufistov/shortygopher.git/internal/app/config"
	"github.com/achufistov/shortygopher.git/internal/app/handlers"
	"github.com/achufistov/shortygopher.git/internal/app/middleware"
	"github.com/achufistov/shortygopher.git/internal/app/storage"
	"github.com/go-chi/chi/v5"
)

func TestInitStorage(t *testing.T) {
	testStorage := storage.NewURLStorage()
	handlers.InitStorage(testStorage)
}

func TestHandlePost_InvalidMethod(t *testing.T) {
	cfg := &config.Config{
		Address:     ":8080",
		BaseURL:     "http://localhost:8080",
		FileStorage: "",
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handlers.HandlePost(cfg, w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandlePost_InvalidContentType(t *testing.T) {
	cfg := &config.Config{
		Address:     ":8080",
		BaseURL:     "http://localhost:8080",
		FileStorage: "",
	}

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("test"))
	req.Header.Set("Content-Type", "application/xml")
	w := httptest.NewRecorder()

	handlers.HandlePost(cfg, w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandlePost_WithValidURL_ReturnsCreatedStatus(t *testing.T) {
	cfg := &config.Config{
		Address:     ":8080",
		BaseURL:     "http://localhost:8080",
		FileStorage: "",
	}

	testStorage := storage.NewURLStorage()
	handlers.InitStorage(testStorage)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com"))
	req.Header.Set("Content-Type", "text/plain")

	// Add user context
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "test-user")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handlers.HandlePost(cfg, w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	if !strings.Contains(w.Body.String(), cfg.BaseURL) {
		t.Fatal("Response should contain base URL")
	}
}

func TestHandleShortenPost_InvalidJSON(t *testing.T) {
	cfg := &config.Config{
		Address:     ":8080",
		BaseURL:     "http://localhost:8080",
		FileStorage: "",
	}

	req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	handlers.HandleShortenPost(cfg, w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandleShortenPost_Success(t *testing.T) {
	cfg := &config.Config{
		Address:     ":8080",
		BaseURL:     "http://localhost:8080",
		FileStorage: "",
	}

	testStorage := storage.NewURLStorage()
	handlers.InitStorage(testStorage)

	reqBody := handlers.ShortenRequest{
		OriginalURL: "https://example.com",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Add user context
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "test-user")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handlers.HandleShortenPost(cfg, w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var resp handlers.ShortenResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !strings.Contains(resp.ShortURL, cfg.BaseURL) {
		t.Fatal("Response should contain base URL")
	}
}

func TestHandleGet_InvalidMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/abc123", nil)
	w := httptest.NewRecorder()

	handlers.HandleGet(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandleGet_NotFound(t *testing.T) {
	testStorage := storage.NewURLStorage()
	handlers.InitStorage(testStorage)

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "nonexistent")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handlers.HandleGet(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestHandleGet_Success(t *testing.T) {
	testStorage := storage.NewURLStorage()
	handlers.InitStorage(testStorage)

	shortURL := "abc123"
	originalURL := "https://example.com"
	userID := "test-user"
	testStorage.AddURL(shortURL, originalURL, userID)

	req := httptest.NewRequest(http.MethodGet, "/"+shortURL, nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", shortURL)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handlers.HandleGet(w, req)

	if w.Code != http.StatusTemporaryRedirect {
		t.Fatalf("Expected status %d, got %d", http.StatusTemporaryRedirect, w.Code)
	}

	if location := w.Header().Get("Location"); location != originalURL {
		t.Fatalf("Expected Location %s, got %s", originalURL, location)
	}
}

func TestHandleGet_Deleted(t *testing.T) {
	testStorage := storage.NewURLStorage()
	handlers.InitStorage(testStorage)

	shortURL := "abc123"
	originalURL := "https://example.com"
	userID := "test-user"
	testStorage.AddURL(shortURL, originalURL, userID)
	testStorage.DeleteURLs([]string{shortURL}, userID)

	req := httptest.NewRequest(http.MethodGet, "/"+shortURL, nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", shortURL)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handlers.HandleGet(w, req)

	if w.Code != http.StatusGone {
		t.Fatalf("Expected status %d, got %d", http.StatusGone, w.Code)
	}
}

func TestHandlePing(t *testing.T) {
	testStorage := storage.NewURLStorage()
	handler := handlers.HandlePing(testStorage)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}
