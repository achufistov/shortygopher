package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/achufistov/shortygopher.git/internal/app/config"
	"github.com/achufistov/shortygopher.git/internal/app/middleware"
	"github.com/achufistov/shortygopher.git/internal/app/storage"
	"github.com/go-chi/chi/v5"
)

func createTestConfig(t *testing.T) *config.Config {
	// Create temporary secret file
	tempDir := t.TempDir()
	secretFile := filepath.Join(tempDir, "secret.key")
	secretContent := "test-secret-key"

	err := os.WriteFile(secretFile, []byte(secretContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test secret file: %v", err)
	}

	// Set environment variables
	os.Setenv("SERVER_ADDRESS", "localhost:8080")
	os.Setenv("BASE_URL", "http://localhost:8080")
	os.Setenv("FILE_STORAGE_PATH", "test_urls.json")
	os.Setenv("JWT_SECRET_FILE", secretFile)

	defer func() {
		os.Unsetenv("SERVER_ADDRESS")
		os.Unsetenv("BASE_URL")
		os.Unsetenv("FILE_STORAGE_PATH")
		os.Unsetenv("JWT_SECRET_FILE")
	}()

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load test config: %v", err)
	}

	return cfg
}

func TestGenerateShortURL(t *testing.T) {
	shortURL1 := generateShortURL()
	shortURL2 := generateShortURL()

	// Check that URLs are generated
	if shortURL1 == "" {
		t.Error("Expected non-empty short URL")
	}
	if shortURL2 == "" {
		t.Error("Expected non-empty short URL")
	}

	// Check that URLs are different (very high probability)
	if shortURL1 == shortURL2 {
		t.Error("Expected different short URLs")
	}

	// Check length (should be 6 characters)
	if len(shortURL1) != 6 {
		t.Errorf("Expected short URL length 6, got %d", len(shortURL1))
	}
	if len(shortURL2) != 6 {
		t.Errorf("Expected short URL length 6, got %d", len(shortURL2))
	}

	// Check that URLs contain only valid base64 URL-safe characters
	validChars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	for _, char := range shortURL1 {
		if !strings.ContainsRune(validChars, char) {
			t.Errorf("Short URL contains invalid character: %c", char)
		}
	}
}

func TestInitStorage(t *testing.T) {
	testStorage := storage.NewURLStorage()

	// Test that InitStorage doesn't panic
	InitStorage(testStorage)

	// Verify that global storage was set (indirect test)
	// We can't directly access storageInstance from tests, but we can verify
	// that subsequent handler calls work correctly
	if testStorage == nil {
		t.Error("Test storage should not be nil")
	}
}

func TestHandlePost_Success(t *testing.T) {
	cfg := createTestConfig(t)
	testStorage := storage.NewURLStorage()
	InitStorage(testStorage)

	req := httptest.NewRequest("POST", "/", strings.NewReader("https://example.com"))
	req.Header.Set("Content-Type", "text/plain")
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "test-user")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	HandlePost(cfg, w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.HasPrefix(body, cfg.BaseURL) {
		t.Errorf("Expected response to start with %s, got %s", cfg.BaseURL, body)
	}
}

func TestHandlePost_InvalidMethod(t *testing.T) {
	cfg := createTestConfig(t)
	testStorage := storage.NewURLStorage()
	InitStorage(testStorage)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	HandlePost(cfg, w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandlePost_Unauthorized(t *testing.T) {
	cfg := createTestConfig(t)
	testStorage := storage.NewURLStorage()
	InitStorage(testStorage)

	req := httptest.NewRequest("POST", "/", strings.NewReader("https://example.com"))
	req.Header.Set("Content-Type", "text/plain")
	// No userID in context
	w := httptest.NewRecorder()

	HandlePost(cfg, w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestHandlePost_InvalidContentType(t *testing.T) {
	cfg := createTestConfig(t)
	testStorage := storage.NewURLStorage()
	InitStorage(testStorage)

	req := httptest.NewRequest("POST", "/", strings.NewReader("https://example.com"))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "test-user")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	HandlePost(cfg, w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleShortenPost_Success(t *testing.T) {
	cfg := createTestConfig(t)
	testStorage := storage.NewURLStorage()
	InitStorage(testStorage)

	reqBody := ShortenRequest{OriginalURL: "https://example.com"}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/shorten", strings.NewReader(string(jsonData)))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "test-user")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	HandleShortenPost(cfg, w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	expectedContentType := "application/json"
	if contentType := w.Header().Get("Content-Type"); contentType != expectedContentType {
		t.Errorf("Expected Content-Type '%s', got '%s'", expectedContentType, contentType)
	}

	var response ShortenResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !strings.HasPrefix(response.ShortURL, cfg.BaseURL) {
		t.Errorf("Expected short URL to start with %s, got %s", cfg.BaseURL, response.ShortURL)
	}
}

func TestHandleShortenPost_InvalidMethod(t *testing.T) {
	cfg := createTestConfig(t)

	req := httptest.NewRequest("GET", "/api/shorten", nil)
	w := httptest.NewRecorder()

	HandleShortenPost(cfg, w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleShortenPost_InvalidJSON(t *testing.T) {
	cfg := createTestConfig(t)
	testStorage := storage.NewURLStorage()
	InitStorage(testStorage)

	req := httptest.NewRequest("POST", "/api/shorten", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "test-user")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	HandleShortenPost(cfg, w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleGet_Success(t *testing.T) {
	testStorage := storage.NewURLStorage()
	InitStorage(testStorage)

	// Add a URL to storage
	testStorage.AddURL("test123", "https://example.com", "user1")

	// Create router to test URL parameter extraction
	r := chi.NewRouter()
	r.Get("/{id}", HandleGet)

	req := httptest.NewRequest("GET", "/test123", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusTemporaryRedirect {
		t.Errorf("Expected status 307, got %d", w.Code)
	}

	location := w.Header().Get("Location")
	if location != "https://example.com" {
		t.Errorf("Expected Location 'https://example.com', got '%s'", location)
	}
}

func TestHandleGet_NotFound(t *testing.T) {
	testStorage := storage.NewURLStorage()
	InitStorage(testStorage)

	r := chi.NewRouter()
	r.Get("/{id}", HandleGet)

	req := httptest.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestHandleGet_InvalidMethod(t *testing.T) {
	testStorage := storage.NewURLStorage()
	InitStorage(testStorage)

	r := chi.NewRouter()
	r.Get("/{id}", HandleGet)

	req := httptest.NewRequest("POST", "/test123", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestHandleGet_DeletedURL(t *testing.T) {
	testStorage := storage.NewURLStorage()
	InitStorage(testStorage)

	// Add and then delete a URL
	testStorage.AddURL("test123", "https://example.com", "user1")
	testStorage.DeleteURLs([]string{"test123"}, "user1")

	r := chi.NewRouter()
	r.Get("/{id}", HandleGet)

	req := httptest.NewRequest("GET", "/test123", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusGone {
		t.Errorf("Expected status 410, got %d", w.Code)
	}
}

func TestHandleBatchShortenPost_Success(t *testing.T) {
	cfg := createTestConfig(t)
	testStorage := storage.NewURLStorage()
	InitStorage(testStorage)

	batchReq := []BatchRequest{
		{CorrelationID: "1", OriginalURL: "https://example.com"},
		{CorrelationID: "2", OriginalURL: "https://google.com"},
	}
	jsonData, _ := json.Marshal(batchReq)

	req := httptest.NewRequest("POST", "/api/shorten/batch", strings.NewReader(string(jsonData)))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "test-user")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	HandleBatchShortenPost(cfg, w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var response []BatchResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response) != 2 {
		t.Errorf("Expected 2 responses, got %d", len(response))
	}

	for _, resp := range response {
		if resp.CorrelationID == "" {
			t.Error("Expected non-empty correlation ID")
		}
		if !strings.HasPrefix(resp.ShortURL, cfg.BaseURL) {
			t.Errorf("Expected short URL to start with %s, got %s", cfg.BaseURL, resp.ShortURL)
		}
	}
}

func TestHandleBatchShortenPost_EmptyBatch(t *testing.T) {
	cfg := createTestConfig(t)
	testStorage := storage.NewURLStorage()
	InitStorage(testStorage)

	emptyBatch := []BatchRequest{}
	jsonData, _ := json.Marshal(emptyBatch)

	req := httptest.NewRequest("POST", "/api/shorten/batch", strings.NewReader(string(jsonData)))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "test-user")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	HandleBatchShortenPost(cfg, w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandlePing_Success(t *testing.T) {
	testStorage := storage.NewURLStorage()
	InitStorage(testStorage)

	handler := HandlePing(testStorage)

	req := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHandlePing_InvalidMethod(t *testing.T) {
	testStorage := storage.NewURLStorage()

	handler := HandlePing(testStorage)

	req := httptest.NewRequest("POST", "/ping", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	expectedBody := "Invalid request method\n"
	if w.Body.String() != expectedBody {
		t.Errorf("Expected body '%s', got '%s'", expectedBody, w.Body.String())
	}
}

func TestHandleGetUserURLs_Unauthorized(t *testing.T) {
	cfg := createTestConfig(t)
	testStorage := storage.NewURLStorage()
	InitStorage(testStorage)

	handler := HandleGetUserURLs(cfg)

	// Request without userID in context
	req := httptest.NewRequest("GET", "/api/user/urls", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestHandleGetUserURLs_NoURLs(t *testing.T) {
	cfg := createTestConfig(t)
	testStorage := storage.NewURLStorage()
	InitStorage(testStorage)

	handler := HandleGetUserURLs(cfg)

	// Request with userID in context but no URLs
	req := httptest.NewRequest("GET", "/api/user/urls", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "test-user")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}
}

func TestHandleGetUserURLs_WithURLs(t *testing.T) {
	cfg := createTestConfig(t)
	testStorage := storage.NewURLStorage()
	InitStorage(testStorage)

	// Add some URLs for the test user
	userID := "test-user"
	testStorage.AddURL("short1", "https://example.com", userID)
	testStorage.AddURL("short2", "https://google.com", userID)

	handler := HandleGetUserURLs(cfg)

	req := httptest.NewRequest("GET", "/api/user/urls", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check Content-Type
	expectedContentType := "application/json"
	if contentType := w.Header().Get("Content-Type"); contentType != expectedContentType {
		t.Errorf("Expected Content-Type '%s', got '%s'", expectedContentType, contentType)
	}

	// Parse response
	var response []struct {
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
	}

	err := json.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response) != 2 {
		t.Errorf("Expected 2 URLs in response, got %d", len(response))
	}

	// Verify URLs are correct
	foundShort1 := false
	foundShort2 := false

	for _, item := range response {
		switch item.OriginalURL {
		case "https://example.com":
			if item.ShortURL != "http://localhost:8080/short1" {
				t.Errorf("Expected short URL 'http://localhost:8080/short1', got '%s'", item.ShortURL)
			}
			foundShort1 = true
		case "https://google.com":
			if item.ShortURL != "http://localhost:8080/short2" {
				t.Errorf("Expected short URL 'http://localhost:8080/short2', got '%s'", item.ShortURL)
			}
			foundShort2 = true
		}
	}

	if !foundShort1 {
		t.Error("Expected to find short1 URL in response")
	}
	if !foundShort2 {
		t.Error("Expected to find short2 URL in response")
	}
}

func TestHandleDeleteUserURLs_InvalidMethod(t *testing.T) {
	cfg := createTestConfig(t)

	handler := HandleDeleteUserURLs(cfg)

	req := httptest.NewRequest("GET", "/api/user/urls", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleDeleteUserURLs_InvalidJSON(t *testing.T) {
	cfg := createTestConfig(t)

	handler := HandleDeleteUserURLs(cfg)

	req := httptest.NewRequest("DELETE", "/api/user/urls", strings.NewReader("invalid json"))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleDeleteUserURLs_ValidRequest(t *testing.T) {
	cfg := createTestConfig(t)
	testStorage := storage.NewURLStorage()
	InitStorage(testStorage)

	handler := HandleDeleteUserURLs(cfg)

	// Prepare JSON array of URLs to delete
	urlsToDelete := []string{"short1", "short2"}
	jsonData, err := json.Marshal(urlsToDelete)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	req := httptest.NewRequest("DELETE", "/api/user/urls", strings.NewReader(string(jsonData)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("Expected status 202, got %d", w.Code)
	}
}

func TestShortenRequest(t *testing.T) {
	// Test ShortenRequest struct
	req := ShortenRequest{
		OriginalURL: "https://example.com",
	}

	if req.OriginalURL != "https://example.com" {
		t.Errorf("Expected OriginalURL 'https://example.com', got '%s'", req.OriginalURL)
	}
}

func TestShortenResponse(t *testing.T) {
	// Test ShortenResponse struct
	resp := ShortenResponse{
		ShortURL: "http://localhost:8080/abc123",
	}

	if resp.ShortURL != "http://localhost:8080/abc123" {
		t.Errorf("Expected ShortURL 'http://localhost:8080/abc123', got '%s'", resp.ShortURL)
	}
}

func TestBatchRequest(t *testing.T) {
	// Test BatchRequest struct
	req := BatchRequest{
		CorrelationID: "123",
		OriginalURL:   "https://example.com",
	}

	if req.CorrelationID != "123" {
		t.Errorf("Expected CorrelationID '123', got '%s'", req.CorrelationID)
	}
	if req.OriginalURL != "https://example.com" {
		t.Errorf("Expected OriginalURL 'https://example.com', got '%s'", req.OriginalURL)
	}
}

func TestBatchResponse(t *testing.T) {
	// Test BatchResponse struct
	resp := BatchResponse{
		CorrelationID: "123",
		ShortURL:      "http://localhost:8080/abc123",
	}

	if resp.CorrelationID != "123" {
		t.Errorf("Expected CorrelationID '123', got '%s'", resp.CorrelationID)
	}
	if resp.ShortURL != "http://localhost:8080/abc123" {
		t.Errorf("Expected ShortURL 'http://localhost:8080/abc123', got '%s'", resp.ShortURL)
	}
}
