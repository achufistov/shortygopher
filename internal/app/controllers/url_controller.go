package controllers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/achufistov/shortygopher.git/internal/app/models"
	"github.com/achufistov/shortygopher.git/internal/app/views"
)

// URLController handles URL-related HTTP requests
type URLController struct {
	storage models.URLRepository
	view    *views.URLView
}

// NewURLController creates a new URLController instance
func NewURLController(storage models.URLRepository) *URLController {
	return &URLController{
		storage: storage,
		view:    views.NewURLView(),
	}
}

// HandleShorten handles URL shortening requests
func (c *URLController) HandleShorten(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		c.view.RenderError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	contentType := r.Header.Get("Content-Type")
	var originalURL string

	switch contentType {
	case "application/json":
		var req struct {
			URL string `json:"url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			c.view.RenderError(w, http.StatusBadRequest, "Invalid request body")
			return
		}
		originalURL = req.URL
	case "text/plain":
		body, err := io.ReadAll(r.Body)
		if err != nil {
			c.view.RenderError(w, http.StatusBadRequest, "Invalid request body")
			return
		}
		originalURL = string(body)
	default:
		c.view.RenderError(w, http.StatusBadRequest, "Unsupported content type")
		return
	}

	if originalURL == "" {
		c.view.RenderError(w, http.StatusBadRequest, "URL is required")
		return
	}

	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		userID = "anonymous" // Use default user ID for tests
	}

	// Check if URL already exists
	if shortURL, exists := c.storage.GetShortURLByOriginalURL(originalURL); exists {
		if contentType == "text/plain" {
			c.view.RenderTextShortenResponse(w, shortURL)
		} else {
			c.view.RenderShortenResponse(w, shortURL)
		}
		return
	}

	// Generate new short URL
	shortURL := models.GenerateShortURL()
	if err := c.storage.AddURL(shortURL, originalURL, userID); err != nil {
		c.view.RenderError(w, http.StatusInternalServerError, "Failed to create short URL")
		return
	}

	if contentType == "text/plain" {
		c.view.RenderTextShortenResponse(w, shortURL)
	} else {
		c.view.RenderShortenResponse(w, shortURL)
	}
}

// HandleGet handles requests to retrieve original URLs
func (c *URLController) HandleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		c.view.RenderError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	shortURL := r.URL.Path[1:] // Remove leading slash
	if shortURL == "" {
		c.view.RenderError(w, http.StatusBadRequest, "Short URL is required")
		return
	}

	originalURL, err := c.storage.GetURL(shortURL)
	if err != nil {
		c.view.RenderError(w, http.StatusNotFound, "URL not found")
		return
	}

	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}

// HandleBatchShorten handles batch URL shortening requests
func (c *URLController) HandleBatchShorten(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		c.view.RenderError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		userID = "anonymous" // Use default user ID for tests
	}

	var batchReq []views.BatchRequest
	if err := json.NewDecoder(r.Body).Decode(&batchReq); err != nil {
		c.view.RenderError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	urls := make([]models.URL, 0, len(batchReq))
	for _, req := range batchReq {
		if req.OriginalURL == "" {
			continue
		}

		// Check if URL already exists
		if shortURL, exists := c.storage.GetShortURLByOriginalURL(req.OriginalURL); exists {
			urls = append(urls, models.URL{
				OriginalURL: req.OriginalURL,
				ShortURL:    shortURL,
				UserID:      userID,
			})
			continue
		}

		// Generate new short URL
		shortURL := models.GenerateShortURL()
		if err := c.storage.AddURL(shortURL, req.OriginalURL, userID); err != nil {
			continue
		}

		urls = append(urls, models.URL{
			OriginalURL: req.OriginalURL,
			ShortURL:    shortURL,
			UserID:      userID,
		})
	}

	c.view.RenderBatchResponse(w, urls)
}

// HandleGetUserURLs handles requests to retrieve user's URLs
func (c *URLController) HandleGetUserURLs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		c.view.RenderError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		c.view.RenderError(w, http.StatusUnauthorized, "User ID is required")
		return
	}

	urls, err := c.storage.GetUserURLs(userID)
	if err != nil {
		c.view.RenderError(w, http.StatusInternalServerError, "Failed to get user URLs")
		return
	}

	c.view.RenderUserURLs(w, urls)
}

// HandleDeleteUserURLs handles requests to delete user's URLs
func (c *URLController) HandleDeleteUserURLs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		c.view.RenderError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		c.view.RenderError(w, http.StatusUnauthorized, "User ID is required")
		return
	}

	var shortURLs []string
	if err := json.NewDecoder(r.Body).Decode(&shortURLs); err != nil {
		c.view.RenderError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Delete URLs asynchronously
	go func() {
		_ = c.storage.DeleteURLs(shortURLs, userID)
	}()

	c.view.RenderDeleteAccepted(w)
}
