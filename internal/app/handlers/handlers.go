package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/achufistov/shortygopher.git/internal/app/config"

	"github.com/go-chi/chi/v5"
)

var (
	URLMap = make(map[string]string)
)

type ShortenRequest struct {
	OriginalURL string `json:"url"`
}

type ShortenResponse struct {
	ShortURL string `json:"result"`
}

func HandlePost(cfg *config.Config, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
		return
	}

	var originalURL string

	// Handle JSON requests
	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		var req ShortenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		originalURL = req.OriginalURL
	} else {
		// Handle plain text requests
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		originalURL = string(body)
	}

	// Generate a short URL and store it
	shortURL := generateShortURL()
	URLMap[shortURL] = originalURL

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%s/%s", cfg.BaseURL, shortURL)
}

func HandleGet(cfg *config.Config, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
		return
	}

	id := chi.URLParam(r, "id")
	originalURL, exists := URLMap[id]

	if !exists {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	// if originalURL == "" {
	// 	http.Error(w, "Invalid URL", http.StatusInternalServerError)
	// 	return
	// }

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
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

	shortURL := generateShortURL()
	URLMap[shortURL] = req.OriginalURL

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

func readBody(r *http.Request) (string, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), "text/plain") {
		return "", errors.New("invalid content type")
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func generateShortURL() string {
	b := make([]byte, 6)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	return base64.URLEncoding.EncodeToString(b)[:6]
}
