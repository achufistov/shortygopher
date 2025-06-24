package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/achufistov/shortygopher.git/internal/app/config"
	"github.com/achufistov/shortygopher.git/internal/app/handlers"
	"github.com/achufistov/shortygopher.git/internal/app/middleware"
	"github.com/achufistov/shortygopher.git/internal/app/storage"
	"github.com/go-chi/chi/v5"
)

var cfg *config.Config

func initConfig() {
	// Создаем временный JWT секрет для тестов
	err := os.MkdirAll("./secrets", 0755)
	if err != nil {
		panic(err)
	}

	secretFile := "./secrets/jwt_secret.key"
	err = os.WriteFile(secretFile, []byte("test-jwt-secret-for-integration-tests"), 0600)
	if err != nil {
		panic(err)
	}

	// Устанавливаем переменную окружения для JWT секрета
	os.Setenv("JWT_SECRET_FILE", secretFile)

	var err2 error
	cfg, err2 = config.LoadConfig()
	if err2 != nil {
		panic(err2)
	}
}

// mockAuthMiddleware добавляет тестовый userID в контекст

func mockAuthMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), middleware.UserIDKey, "test_user")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Test_handlePost(t *testing.T) {

	initConfig()
	storageInstance := storage.NewURLStorage()
	handlers.InitStorage(storageInstance)
	tests := []struct {
		name           string
		requestBody    string
		contentType    string
		expectedStatus int
		expectedURL    string
	}{
		{
			name:           "Valid POST request (text/plain)",
			requestBody:    "https://example.com",
			contentType:    "text/plain",
			expectedStatus: http.StatusCreated,
			expectedURL:    cfg.BaseURL + "/",
		},
		{
			name:           "Invalid content type",
			requestBody:    "https://example.com",
			contentType:    "application/xml",
			expectedStatus: http.StatusBadRequest,
			expectedURL:    "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/", bytes.NewBufferString(tt.requestBody))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", tt.contentType)
			rr := httptest.NewRecorder()
			handler := mockAuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlers.HandlePost(cfg, w, r)
			}))
			handler.ServeHTTP(rr, req)
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}
			if tt.expectedStatus == http.StatusCreated {
				if !strings.HasPrefix(rr.Body.String(), tt.expectedURL) {
					t.Errorf("handler returned unexpected body: got %v want %v",
						rr.Body.String(), tt.expectedURL)
				}
			}
		})
	}
}

func Test_handleGet(t *testing.T) {

	initConfig()
	storageInstance := storage.NewURLStorage()
	handlers.InitStorage(storageInstance)
	shortURL := "abc123"
	originalURL := "https://example.com"
	userID := "test_user"
	storageInstance.AddURL(shortURL, originalURL, userID) // Добавляем userID
	tests := []struct {
		name           string
		urlPath        string
		expectedStatus int
		expectedURL    string
	}{
		{
			name:           "Valid GET request",
			urlPath:        "/" + shortURL,
			expectedStatus: http.StatusTemporaryRedirect,
			expectedURL:    originalURL,
		},
		{
			name:           "Invalid GET request",
			urlPath:        "/nonexistent",
			expectedStatus: http.StatusNotFound,
			expectedURL:    "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.urlPath, nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			r := chi.NewRouter()
			r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
				handlers.HandleGet(w, r)
			})
			r.ServeHTTP(rr, req)
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}
			if tt.expectedStatus == http.StatusTemporaryRedirect {
				if location := rr.Header().Get("Location"); location != tt.expectedURL {
					t.Errorf("handler returned unexpected Location header: got %v want %v",
						location, tt.expectedURL)
				}
			}
		})
	}
}

func Test_handleShortenPost(t *testing.T) {

	initConfig()
	storageInstance := storage.NewURLStorage()
	handlers.InitStorage(storageInstance)
	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		expectedJSON   string
	}{
		{
			name:           "Valid POST request",
			requestBody:    `{"url": "https://example.com"}`,
			expectedStatus: http.StatusCreated,
			expectedJSON:   `{"result":"` + cfg.BaseURL + `/`,
		},
		{
			name:           "Invalid JSON",
			requestBody:    `{"url": "invalid-url"`,
			expectedStatus: http.StatusBadRequest,
			expectedJSON:   "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/api/shorten", bytes.NewBufferString(tt.requestBody))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			handler := mockAuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlers.HandleShortenPost(cfg, w, r)
			}))
			handler.ServeHTTP(rr, req)
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}
			if tt.expectedStatus == http.StatusCreated {
				if !strings.HasPrefix(rr.Body.String(), tt.expectedJSON) {
					t.Errorf("handler returned unexpected body: got %v want %v",
						rr.Body.String(), tt.expectedJSON)
				}
			}
		})
	}
}

func Test_handlePing(t *testing.T) {

	initConfig()
	storageInstance := storage.NewURLStorage()
	handlers.InitStorage(storageInstance)
	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "Valid GET request",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid request method (POST)",
			method:         http.MethodPost,
			expectedStatus: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, "/ping", nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlers.HandlePing(storageInstance)(w, r)
			})
			handler.ServeHTTP(rr, req)
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}
		})
	}
}

func Test_handleBatchShortenPost(t *testing.T) {

	initConfig()
	storageInstance := storage.NewURLStorage()
	handlers.InitStorage(storageInstance)
	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
	}{
		{
			name:           "Invalid JSON",
			requestBody:    `[{"correlation_id": "1", "original_url": "invalid-url"`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty batch",
			requestBody:    `[]`,
			expectedStatus: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/api/shorten/batch", bytes.NewBufferString(tt.requestBody))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			handler := mockAuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlers.HandleBatchShortenPost(cfg, w, r)
			}))
			handler.ServeHTTP(rr, req)
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}
		})
	}
}
