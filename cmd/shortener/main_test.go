package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/achufistov/shortygopher.git/internal/app/config"
	"github.com/achufistov/shortygopher.git/internal/app/controllers"
	"github.com/achufistov/shortygopher.git/internal/app/middleware"
	"github.com/achufistov/shortygopher.git/internal/app/storage"
)

var (
	cfg *config.Config
)

func initConfig() {
	var err error
	cfg, err = config.LoadConfig()
	if err != nil {
		panic(err)
	}
}

// mockAuthMiddleware adds test userID to context
func mockAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), middleware.UserIDKey, "test_user")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Test_handleShorten(t *testing.T) {
	initConfig()
	storageInstance := storage.NewURLStorage()
	urlController := controllers.NewURLController(storageInstance)

	// Add test URL
	originalURL := "https://example.com"
	shortURL := "abc123"
	userID := "test_user"
	storageInstance.AddURL(shortURL, originalURL, userID)

	tests := []struct {
		name           string
		method         string
		requestBody    string
		contentType    string
		expectedStatus int
		expectedURL    string
		expectedError  string
	}{
		{
			name:           "Valid POST request (text/plain)",
			method:         http.MethodPost,
			requestBody:    "https://example.com",
			contentType:    "text/plain",
			expectedStatus: http.StatusCreated,
			expectedURL:    shortURL,
		},
		{
			name:           "Valid POST request (application/json)",
			method:         http.MethodPost,
			requestBody:    `{"url": "https://example.com"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusCreated,
			expectedURL:    shortURL,
		},
		{
			name:           "Invalid content type",
			method:         http.MethodPost,
			requestBody:    "https://example.com",
			contentType:    "application/xml",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Unsupported content type",
		},
		{
			name:           "Empty URL",
			method:         http.MethodPost,
			requestBody:    "",
			contentType:    "text/plain",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "URL is required",
		},
		{
			name:           "Method not allowed",
			method:         http.MethodGet,
			requestBody:    "https://example.com",
			contentType:    "text/plain",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedError:  "Method not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, "/", bytes.NewBufferString(tt.requestBody))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", tt.contentType)
			rr := httptest.NewRecorder()
			handler := mockAuthMiddleware(http.HandlerFunc(urlController.HandleShorten))
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}

			if tt.expectedStatus == http.StatusCreated {
				if tt.contentType == "text/plain" {
					if !strings.Contains(rr.Body.String(), tt.expectedURL) {
						t.Errorf("handler returned unexpected body: got %v want %v",
							rr.Body.String(), tt.expectedURL)
					}
				} else {
					var response struct {
						Result string `json:"result"`
					}
					if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
						t.Errorf("failed to decode response: %v", err)
					}
					if !strings.Contains(response.Result, tt.expectedURL) {
						t.Errorf("handler returned unexpected body: got %v want %v",
							response.Result, tt.expectedURL)
					}
				}
			} else if tt.expectedError != "" {
				var response struct {
					Error string `json:"error"`
				}
				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Errorf("failed to decode error response: %v", err)
				}
				if response.Error != tt.expectedError {
					t.Errorf("handler returned unexpected error: got %v want %v",
						response.Error, tt.expectedError)
				}
			}
		})
	}
}

func Test_handleGet(t *testing.T) {
	initConfig()
	storageInstance := storage.NewURLStorage()
	urlController := controllers.NewURLController(storageInstance)

	shortURL := "abc123"
	originalURL := "https://example.com"
	userID := "test_user"
	storageInstance.AddURL(shortURL, originalURL, userID)

	tests := []struct {
		name           string
		method         string
		urlPath        string
		expectedStatus int
		expectedURL    string
		expectedError  string
	}{
		{
			name:           "Valid GET request",
			method:         http.MethodGet,
			urlPath:        "/" + shortURL,
			expectedStatus: http.StatusTemporaryRedirect,
			expectedURL:    originalURL,
		},
		{
			name:           "Invalid GET request - nonexistent URL",
			method:         http.MethodGet,
			urlPath:        "/nonexistent",
			expectedStatus: http.StatusNotFound,
			expectedError:  "URL not found",
		},
		{
			name:           "Method not allowed",
			method:         http.MethodPost,
			urlPath:        "/" + shortURL,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedError:  "Method not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, tt.urlPath, nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			urlController.HandleGet(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}

			if tt.expectedStatus == http.StatusTemporaryRedirect {
				if location := rr.Header().Get("Location"); location != tt.expectedURL {
					t.Errorf("handler returned unexpected Location header: got %v want %v",
						location, tt.expectedURL)
				}
			} else if tt.expectedError != "" {
				var response struct {
					Error string `json:"error"`
				}
				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Errorf("failed to decode error response: %v", err)
				}
				if response.Error != tt.expectedError {
					t.Errorf("handler returned unexpected error: got %v want %v",
						response.Error, tt.expectedError)
				}
			}
		})
	}
}

func Test_handleBatchShorten(t *testing.T) {
	initConfig()
	storageInstance := storage.NewURLStorage()
	urlController := controllers.NewURLController(storageInstance)

	// Add test URL
	originalURL := "https://example.com"
	shortURL := "abc123"
	userID := "test_user"
	storageInstance.AddURL(shortURL, originalURL, userID)

	tests := []struct {
		name           string
		method         string
		requestBody    string
		contentType    string
		expectedStatus int
		expectedURLs   []string
		expectedError  string
	}{
		{
			name:           "Valid batch request with existing URL",
			method:         http.MethodPost,
			requestBody:    `[{"correlation_id": "1", "original_url": "https://example.com"}]`,
			contentType:    "application/json",
			expectedStatus: http.StatusCreated,
			expectedURLs:   []string{shortURL},
		},
		{
			name:           "Invalid request body",
			method:         http.MethodPost,
			requestBody:    `invalid json`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request body",
		},
		{
			name:           "Method not allowed",
			method:         http.MethodGet,
			requestBody:    `[{"correlation_id": "1", "original_url": "https://example.com"}]`,
			contentType:    "application/json",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedError:  "Method not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, "/api/shorten/batch", bytes.NewBufferString(tt.requestBody))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", tt.contentType)
			rr := httptest.NewRecorder()
			handler := mockAuthMiddleware(http.HandlerFunc(urlController.HandleBatchShorten))
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}

			if tt.expectedStatus == http.StatusCreated {
				var response []struct {
					OriginalURL string `json:"original_url"`
					ShortURL    string `json:"short_url"`
				}
				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Errorf("failed to decode response: %v", err)
				}
				for i, expectedURL := range tt.expectedURLs {
					if !strings.Contains(response[i].ShortURL, expectedURL) {
						t.Errorf("handler returned unexpected URL: got %v want %v",
							response[i].ShortURL, expectedURL)
					}
				}
			} else if tt.expectedError != "" {
				var response struct {
					Error string `json:"error"`
				}
				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Errorf("failed to decode error response: %v", err)
				}
				if response.Error != tt.expectedError {
					t.Errorf("handler returned unexpected error: got %v want %v",
						response.Error, tt.expectedError)
				}
			}
		})
	}
}

func Test_handleGetUserURLs(t *testing.T) {
	initConfig()
	storageInstance := storage.NewURLStorage()
	urlController := controllers.NewURLController(storageInstance)

	// Add test URLs
	urls := []struct {
		shortURL    string
		originalURL string
	}{
		{"abc123", "https://example1.com"},
		{"def456", "https://example2.com"},
	}

	userID := "test_user"
	for _, u := range urls {
		storageInstance.AddURL(u.shortURL, u.originalURL, userID)
	}

	tests := []struct {
		name           string
		method         string
		userID         string
		expectedStatus int
		expectedURLs   int
		expectedError  string
	}{
		{
			name:           "Valid GET request",
			method:         http.MethodGet,
			userID:         userID,
			expectedStatus: http.StatusOK,
			expectedURLs:   2,
		},
		{
			name:           "Method not allowed",
			method:         http.MethodPost,
			userID:         userID,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedError:  "Method not allowed",
		},
		{
			name:           "Missing user ID",
			method:         http.MethodGet,
			userID:         "",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "User ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, "/api/user/urls", nil)
			if err != nil {
				t.Fatal(err)
			}
			if tt.userID != "" {
				req.Header.Set("X-User-Id", tt.userID)
			}
			rr := httptest.NewRecorder()
			urlController.HandleGetUserURLs(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}

			if tt.expectedStatus == http.StatusOK {
				var response []struct {
					OriginalURL string `json:"original_url"`
					ShortURL    string `json:"short_url"`
				}
				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Errorf("failed to decode response: %v", err)
				}
				if len(response) != tt.expectedURLs {
					t.Errorf("handler returned wrong number of URLs: got %v want %v",
						len(response), tt.expectedURLs)
				}
			} else if tt.expectedError != "" {
				var response struct {
					Error string `json:"error"`
				}
				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Errorf("failed to decode error response: %v", err)
				}
				if response.Error != tt.expectedError {
					t.Errorf("handler returned unexpected error: got %v want %v",
						response.Error, tt.expectedError)
				}
			}
		})
	}
}

func Test_handleDeleteUserURLs(t *testing.T) {
	initConfig()
	storageInstance := storage.NewURLStorage()
	urlController := controllers.NewURLController(storageInstance)

	// Add test URLs
	urls := []struct {
		shortURL    string
		originalURL string
	}{
		{"abc123", "https://example1.com"},
		{"def456", "https://example2.com"},
	}

	userID := "test_user"
	for _, u := range urls {
		storageInstance.AddURL(u.shortURL, u.originalURL, userID)
	}

	tests := []struct {
		name           string
		method         string
		userID         string
		requestBody    string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Valid DELETE request",
			method:         http.MethodDelete,
			userID:         userID,
			requestBody:    `["abc123", "def456"]`,
			expectedStatus: http.StatusAccepted,
		},
		{
			name:           "Method not allowed",
			method:         http.MethodPost,
			userID:         userID,
			requestBody:    `["abc123"]`,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedError:  "Method not allowed",
		},
		{
			name:           "Missing user ID",
			method:         http.MethodDelete,
			userID:         "",
			requestBody:    `["abc123"]`,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "User ID is required",
		},
		{
			name:           "Invalid request body",
			method:         http.MethodDelete,
			userID:         userID,
			requestBody:    `invalid json`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, "/api/user/urls", bytes.NewBufferString(tt.requestBody))
			if err != nil {
				t.Fatal(err)
			}
			if tt.userID != "" {
				req.Header.Set("X-User-Id", tt.userID)
			}
			rr := httptest.NewRecorder()
			urlController.HandleDeleteUserURLs(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}

			if tt.expectedStatus == http.StatusAccepted {
				var response struct {
					Message string `json:"message"`
				}
				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Errorf("failed to decode response: %v", err)
				}
				if response.Message != "URL deletion request accepted" {
					t.Errorf("handler returned unexpected message: got %v want %v",
						response.Message, "URL deletion request accepted")
				}
			} else if tt.expectedError != "" {
				var response struct {
					Error string `json:"error"`
				}
				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Errorf("failed to decode error response: %v", err)
				}
				if response.Error != tt.expectedError {
					t.Errorf("handler returned unexpected error: got %v want %v",
						response.Error, tt.expectedError)
				}
			}
		})
	}
}
