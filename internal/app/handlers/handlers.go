// Package handlers provides HTTP handlers for the URL shortening service.
package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/achufistov/shortygopher.git/internal/app/config"
	"github.com/achufistov/shortygopher.git/internal/app/middleware"
	"github.com/achufistov/shortygopher.git/internal/app/storage"
	"github.com/go-chi/chi/v5"
)

var storageInstance storage.Storage

// ShortenRequest represents a URL shortening request in JSON format.
// Used in the POST /api/shorten endpoint.
//
// Example JSON:
//
//	{
//	  "url": "https://example.com/very/long/path"
//	}
type ShortenRequest struct {
	OriginalURL string `json:"url"`
}

// ShortenResponse represents a URL shortening response in JSON format.
// Returned from the POST /api/shorten endpoint.
//
// Example JSON:
//
//	{
//	  "result": "http://localhost:8080/abc123"
//	}
type ShortenResponse struct {
	ShortURL string `json:"result"`
}

// BatchRequest represents one item in a batch request for shortening multiple URLs.
// Used in the POST /api/shorten/batch endpoint.
type BatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// BatchResponse represents one item in a batch response for shortening multiple URLs.
// Returned from the POST /api/shorten/batch endpoint.
type BatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// StatsResponse represents the response from the GET /api/internal/stats endpoint.
// Contains statistics about the URL shortening service.
type StatsResponse struct {
	URLs  int `json:"urls"`
	Users int `json:"users"`
}

// InitStorage initializes the global storage instance.
// Must be called before using any handlers.
//
// Example:
//
//	storage := storage.NewURLStorage()
//	handlers.InitStorage(storage)
func InitStorage(storage storage.Storage) {
	storageInstance = storage
}

// HandlePost handles POST / requests for URL shortening in text format.
// Accepts the original URL in the request body as text/plain.
// Returns the shortened URL in the response body.
//
// HTTP methods: POST
// Content-Type: text/plain
// Response: text/plain with shortened URL
//
// Response codes:
//   - 201: URL successfully shortened
//   - 400: Invalid request method or Content-Type
//   - 401: User not authorized
//   - 409: URL already exists
//   - 500: Internal server error
func HandlePost(cfg *config.Config, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
		return
	}

	var originalURL string

	contentType := r.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") && !strings.Contains(contentType, "text/plain") {
		http.Error(w, "Invalid content type", http.StatusBadRequest)
		return
	}

	if strings.Contains(contentType, "application/json") {
		var req ShortenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		originalURL = req.OriginalURL
	} else {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		originalURL = string(body)
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	shortURL := generateShortURL()
	err := storageInstance.AddURL(shortURL, originalURL, userID)
	if err != nil {
		if err.Error() == "URL already exists" {
			existingShortURL, exists := storageInstance.GetShortURLByOriginalURL(originalURL)
			if !exists {
				http.Error(w, "Failed to get existing short URL", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusConflict)
			fmt.Fprintf(w, "%s/%s", cfg.BaseURL, existingShortURL)
			return
		}
		http.Error(w, "Failed to save URL mapping", http.StatusInternalServerError)
		return
	}

	if cfg.FileStorage != "" {
		if err := storage.SaveSingleURLMapping(cfg.FileStorage, shortURL, originalURL); err != nil {
			log.Printf("Warning: Failed to save URL mapping to file: %v", err)
		}
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%s/%s", cfg.BaseURL, shortURL)
}

// HandleShortenPost handles POST /api/shorten requests for URL shortening in JSON format.
// Accepts JSON with original URL and returns JSON with shortened URL.
//
// HTTP methods: POST
// Content-Type: application/json
// Response: application/json with ShortenResponse object
//
// Response codes:
//   - 201: URL successfully shortened
//   - 400: Invalid request method or JSON
//   - 401: User not authorized
//   - 409: URL already exists
//   - 500: Internal server error
func HandleShortenPost(cfg *config.Config, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
		return
	}

	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	shortURL := generateShortURL()
	err := storageInstance.AddURL(shortURL, req.OriginalURL, userID)
	if err != nil {
		if err.Error() == "URL already exists" {
			existingShortURL, exists := storageInstance.GetShortURLByOriginalURL(req.OriginalURL)
			if !exists {
				http.Error(w, "Failed to get existing short URL", http.StatusInternalServerError)
				return
			}
			resp := ShortenResponse{
				ShortURL: fmt.Sprintf("%s/%s", cfg.BaseURL, existingShortURL),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			}
			return
		}
		http.Error(w, "Failed to save URL mapping", http.StatusInternalServerError)
		return
	}

	if cfg.FileStorage != "" {
		if err := storage.SaveSingleURLMapping(cfg.FileStorage, shortURL, req.OriginalURL); err != nil {
			log.Printf("Warning: Failed to save URL mapping to file: %v", err)
		}
	}

	resp := ShortenResponse{
		ShortURL: fmt.Sprintf("%s/%s", cfg.BaseURL, shortURL),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// HandleGet handles GET /{id} requests for redirecting to the original URL.
// Looks up the original URL by short identifier and performs HTTP redirect.
//
// HTTP methods: GET
// URL parameters: id - short URL identifier
// Response: HTTP redirect (307 Temporary Redirect)
//
// Response codes:
//   - 307: Successful redirect to original URL
//   - 400: Invalid request method
//   - 404: URL not found
//   - 410: URL was deleted
func HandleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
		return
	}

	id := chi.URLParam(r, "id")
	originalURL, exists, isDeleted := storageInstance.GetURL(id)

	if !exists {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	if isDeleted {
		http.Error(w, "URL has been deleted", http.StatusGone)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

// HandlePing returns a handler for checking storage availability.
// The endpoint is used for health checks and monitoring.
//
// HTTP methods: GET
// URL: /ping
// Response: HTTP status without body
//
// Response codes:
//   - 200: Storage is available
//   - 400: Invalid request method
//   - 500: Storage is unavailable
func HandlePing(storageInstance storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Invalid request method", http.StatusBadRequest)
			return
		}
		if err := storageInstance.Ping(); err != nil {
			http.Error(w, "Failed to ping storage", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

// HandleBatchShortenPost handles POST /api/shorten/batch requests for shortening multiple URLs at once.
// Accepts an array of BatchRequest and returns an array of BatchResponse with shortened URLs.
//
// HTTP methods: POST
// Content-Type: application/json
// Response: application/json with array of BatchResponse
//
// Example request:
//
//	[
//	  {"correlation_id": "1", "original_url": "https://example.com"},
//	  {"correlation_id": "2", "original_url": "https://google.com"}
//	]
//
// Response codes:
//   - 201: URLs successfully shortened
//   - 400: Invalid request method, JSON, or empty array
//   - 401: User not authorized
//   - 500: Internal server error
func HandleBatchShortenPost(cfg *config.Config, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
		return
	}
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var batchRequests []BatchRequest
	if err := json.NewDecoder(r.Body).Decode(&batchRequests); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if len(batchRequests) == 0 {
		http.Error(w, "Empty batch", http.StatusBadRequest)
		return
	}

	batchResponses := make([]BatchResponse, 0, len(batchRequests))

	urlsToSave := make(map[string]string, len(batchRequests))

	for _, req := range batchRequests {
		shortURL := generateShortURL()
		err := storageInstance.AddURL(shortURL, req.OriginalURL, userID)
		if err != nil && err.Error() != "URL already exists" {
			http.Error(w, "Failed to save URL mapping", http.StatusInternalServerError)
			return
		}
		batchResponses = append(batchResponses, BatchResponse{
			CorrelationID: req.CorrelationID,
			ShortURL:      fmt.Sprintf("%s/%s", cfg.BaseURL, shortURL),
		})
		urlsToSave[shortURL] = req.OriginalURL
	}

	if cfg.FileStorage != "" && len(urlsToSave) > 0 {
		if err := storage.SaveURLMappings(cfg.FileStorage, urlsToSave); err != nil {
			log.Printf("Warning: Failed to save URL mappings to file: %v", err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(batchResponses); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// HandleGetUserURLs returns a handler for getting all URLs created by the authenticated user.
// Requires user authentication via JWT token in cookies.
//
// HTTP methods: GET
// Content-Type: application/json
// Response: JSON array of user's URLs with short_url and original_url fields
//
// Response codes:
//   - 200: URLs successfully retrieved
//   - 204: User has no URLs
//   - 401: User not authenticated
//   - 500: Internal server error
func HandleGetUserURLs(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(middleware.UserIDKey).(string)
		if !ok || userID == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		urls, err := storageInstance.GetURLsByUser(userID)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if len(urls) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		response := make([]struct {
			ShortURL    string `json:"short_url"`
			OriginalURL string `json:"original_url"`
		}, 0)
		for short, original := range urls {
			response = append(response, struct {
				ShortURL    string `json:"short_url"`
				OriginalURL string `json:"original_url"`
			}{
				ShortURL:    fmt.Sprintf("%s/%s", cfg.BaseURL, short),
				OriginalURL: original,
			})
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}

// HandleDeleteUserURLs returns a handler for asynchronously deleting specified URLs.
// Accepts a JSON array of short URL IDs and marks them for deletion.
//
// HTTP methods: DELETE
// Content-Type: application/json
// Request body: JSON array of short URL strings
//
// Response codes:
//   - 202: Deletion request accepted (async operation)
//   - 400: Invalid request method or JSON body
func HandleDeleteUserURLs(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "Invalid request method", http.StatusBadRequest)
			return
		}

		var shortURLs []string
		if err := json.NewDecoder(r.Body).Decode(&shortURLs); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		deleteChan := make(chan error)

		go func() {
			err := storageInstance.DeleteURLs(shortURLs, "")
			deleteChan <- err
		}()

		w.WriteHeader(http.StatusAccepted)

		go func() {
			err := <-deleteChan
			if err != nil {
				log.Printf("Failed to delete URLs: %v", err)
			} else {
				log.Println("URLs deleted successfully")
			}
		}()
	}
}

// HandleGetStats handles GET /api/internal/stats requests for service statistics.
// Returns the number of URLs and unique users in the service.
// Access is restricted to clients within the trusted subnet.
//
// HTTP methods: GET
// Response: application/json with statistics
//
// Response codes:
//   - 200: Statistics retrieved successfully
//   - 403: Access denied (not in trusted subnet)
//   - 500: Internal server error
func HandleGetStats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		urlCount, userCount, err := storageInstance.GetStats()
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		response := StatsResponse{
			URLs:  urlCount,
			Users: userCount,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}

func generateShortURL() string {
	b := make([]byte, 6)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	return base64.URLEncoding.EncodeToString(b)[:6]
}
