package handlers

import (
	"crypto/rand"
	"encoding/base64"
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

func HandlePost(cfg *config.Config, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
		return
	}

	originalURL, err := readBody(r)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

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

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
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
