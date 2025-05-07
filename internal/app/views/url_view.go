package views

import (
	"encoding/json"
	"net/http"

	"github.com/achufistov/shortygopher.git/internal/app/models"
)

// URLView handles the presentation logic for URL-related responses
type URLView struct{}

// NewURLView creates a new URLView instance
func NewURLView() *URLView {
	return &URLView{}
}

// RenderShortenResponse sends a JSON response with the shortened URL
func (v *URLView) RenderShortenResponse(w http.ResponseWriter, shortURL string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"result": shortURL,
	})
}

// RenderTextShortenResponse sends a plain text response with the shortened URL
func (v *URLView) RenderTextShortenResponse(w http.ResponseWriter, shortURL string) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

// RenderError sends a JSON error response
func (v *URLView) RenderError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

// RenderBatchResponse sends a JSON response with multiple shortened URLs
func (v *URLView) RenderBatchResponse(w http.ResponseWriter, urls []models.URL) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(urls)
}

// RenderUserURLs sends a JSON response with user's URLs
func (v *URLView) RenderUserURLs(w http.ResponseWriter, urls []models.URL) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(urls)
}

// RenderDeleteAccepted sends a JSON response confirming URL deletion request
func (v *URLView) RenderDeleteAccepted(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "URL deletion request accepted",
	})
}

// BatchRequest represents a request for batch URL shortening
type BatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// BatchResponse represents a response for batch URL shortening
type BatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
