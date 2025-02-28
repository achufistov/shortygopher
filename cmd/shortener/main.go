package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/achufistov/shortygopher.git/cmd/shortener/config"

	"github.com/go-chi/chi/v5"
)

var (
	urlMap = make(map[string]string)
	cfg    *config.Config
)

func main() {
	var err error
	cfg, err = config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	r := chi.NewRouter()

	r.Post("/", handlePost)
	r.Get("/{id}", handleGet)

	log.Printf("Server is running on %s", cfg.Address)
	log.Fatal(http.ListenAndServe(cfg.Address, r))
}

func handlePost(w http.ResponseWriter, r *http.Request) {
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
	urlMap[shortURL] = originalURL

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%s/%s", cfg.BaseURL, shortURL)
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
		return
	}

	id := chi.URLParam(r, "id")
	originalURL, exists := urlMap[id]

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
