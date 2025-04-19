package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/achufistov/shortygopher.git/internal/app/config"
	"github.com/achufistov/shortygopher.git/internal/app/handlers"
	"github.com/achufistov/shortygopher.git/internal/app/storage"

	"github.com/go-chi/chi/v5"
)

func initConfig() {
	var err error
	cfg, err = config.LoadConfig()
	if err != nil {
		panic(err)
	}
}

func Test_handlePost(t *testing.T) {
	initConfig()

	// init in-memory storage
	storageInstance := storage.NewURLStorage()
	handlers.InitStorage(storageInstance)

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		expectedURL    string
	}{
		{
			name:           "Valid POST request",
			requestBody:    "https://example.com",
			expectedStatus: http.StatusCreated,
			expectedURL:    cfg.BaseURL + "/",
		},
		{
			name:           "Invalid content type",
			requestBody:    "https://example.com",
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

			if tt.expectedStatus == http.StatusCreated {
				req.Header.Set("Content-Type", "text/plain")
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlers.HandlePost(cfg, w, r)
			})

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
	storageInstance.AddURL(shortURL, originalURL)

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
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlers.HandleShortenPost(cfg, w, r)
			})

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
		expectedStatus int
	}{
		{
			name:           "Valid GET request",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid request method",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.expectedStatus == http.StatusOK {
				var err error
				req, err = http.NewRequest("GET", "/ping", nil)
				if err != nil {
					t.Fatal(err)
				}
			} else {
				var err error
				req, err = http.NewRequest("POST", "/ping", nil)
				if err != nil {
					t.Fatal(err)
				}
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
		expectedJSON   string
	}{
		{
			name:           "Invalid JSON",
			requestBody:    `[{"correlation_id": "1", "original_url": "invalid-url"`,
			expectedStatus: http.StatusBadRequest,
			expectedJSON:   "",
		},
		{
			name:           "Empty batch",
			requestBody:    `[]`,
			expectedStatus: http.StatusBadRequest,
			expectedJSON:   "",
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
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlers.HandleBatchShortenPost(cfg, w, r)
			})

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
