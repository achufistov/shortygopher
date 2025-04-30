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

var (
	storageInstance storage.Storage
)

type ShortenRequest struct {
	OriginalURL string `json:"url"`
}

type ShortenResponse struct {
	ShortURL string `json:"result"`
}

type BatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

func InitStorage(storage storage.Storage) {
	storageInstance = storage
}

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

	if err := storage.SaveURLMappings(cfg.FileStorage, storageInstance.GetAllURLs()); err != nil {
		http.Error(w, "Failed to save URL mapping", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%s/%s", cfg.BaseURL, shortURL)
}

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
	if err := storage.SaveURLMappings(cfg.FileStorage, storageInstance.GetAllURLs()); err != nil {
		http.Error(w, "Failed to save URL mapping", http.StatusInternalServerError)
		return
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

func HandleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
		return
	}

	id := chi.URLParam(r, "id")
	originalURL, exists := storageInstance.GetURL(id)

	if !exists {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

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
	}
	if err := storage.SaveURLMappings(cfg.FileStorage, storageInstance.GetAllURLs()); err != nil {
		http.Error(w, "Failed to save URL mappings", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(batchResponses); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

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

		if err := storageInstance.DeleteURLs(shortURLs, ""); err != nil {
			http.Error(w, "Failed to delete URLs", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
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
