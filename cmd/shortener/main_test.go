package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/achufistov/shortygopher.git/internal/app/config"
	"github.com/achufistov/shortygopher.git/internal/app/handlers"

	"github.com/go-chi/chi/v5"
)

func initConfig() {
	var err error
	cfg, err = config.LoadConfig()
	if err != nil {
		panic(err) // nil pointer dereference pass the fourth test
	}
}

func Test_handlePost(t *testing.T) {
	initConfig()

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
			expectedURL:    cfg.BaseURL + "/", // use cfg.BaseURL to pass the fifth test
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
	// Prepopulate the urlMap with a test URL
	shortURL := "abc123"
	originalURL := "https://example.com"
	handlers.URLMap[shortURL] = originalURL

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
			r := chi.NewRouter() // to fix this problem, I used the chi router in this test so that the URL parameters were extracted correctly
			r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
				handlers.HandleGet(cfg, w, r)
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
