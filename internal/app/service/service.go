// Package service provides business logic for the URL shortening service.
package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"

	"github.com/achufistov/shortygopher.git/internal/app/config"
	"github.com/achufistov/shortygopher.git/internal/app/storage"
)

// Service provides business logic for URL shortening operations.
type Service struct {
	storage storage.Storage
	config  *config.Config
}

// NewService creates a new service instance.
func NewService(storage storage.Storage, config *config.Config) *Service {
	return &Service{
		storage: storage,
		config:  config,
	}
}

// ShortenURLRequest represents a request to shorten a URL.
type ShortenURLRequest struct {
	OriginalURL string
	UserID      string
}

// ShortenURLResponse represents a response with shortened URL.
type ShortenURLResponse struct {
	ShortURL     string
	AlreadyExists bool
}

// ShortenURL creates a shortened URL from the original URL.
func (s *Service) ShortenURL(ctx context.Context, req ShortenURLRequest) (*ShortenURLResponse, error) {
	shortURL := generateShortURL()
	err := s.storage.AddURL(shortURL, req.OriginalURL, req.UserID)
	
	if err != nil {
		if err.Error() == "URL already exists" {
			existingShortURL, exists := s.storage.GetShortURLByOriginalURL(req.OriginalURL)
			if !exists {
				return nil, fmt.Errorf("failed to get existing short URL")
			}
			return &ShortenURLResponse{
				ShortURL:     fmt.Sprintf("%s/%s", s.config.BaseURL, existingShortURL),
				AlreadyExists: true,
			}, nil
		}
		return nil, fmt.Errorf("failed to save URL mapping: %w", err)
	}

	// Save to file if configured
	if s.config.FileStorage != "" {
		if err := storage.SaveSingleURLMapping(s.config.FileStorage, shortURL, req.OriginalURL); err != nil {
			log.Printf("Warning: Failed to save URL mapping to file: %v", err)
		}
	}

	return &ShortenURLResponse{
		ShortURL:     fmt.Sprintf("%s/%s", s.config.BaseURL, shortURL),
		AlreadyExists: false,
	}, nil
}

// GetURLRequest represents a request to get original URL by short ID.
type GetURLRequest struct {
	ShortID string
}

// GetURLResponse represents a response with original URL.
type GetURLResponse struct {
	OriginalURL string
	Exists      bool
	Deleted     bool
}

// GetURL retrieves the original URL by short ID.
func (s *Service) GetURL(ctx context.Context, req GetURLRequest) (*GetURLResponse, error) {
	originalURL, exists, isDeleted := s.storage.GetURL(req.ShortID)
	
	return &GetURLResponse{
		OriginalURL: originalURL,
		Exists:      exists,
		Deleted:     isDeleted,
	}, nil
}

// BatchRequest represents one item in a batch request.
type BatchRequest struct {
	CorrelationID string
	OriginalURL   string
}

// BatchResponse represents one item in a batch response.
type BatchResponse struct {
	CorrelationID string
	ShortURL      string
}

// ShortenURLBatchRequest represents a request to shorten multiple URLs.
type ShortenURLBatchRequest struct {
	URLs   []BatchRequest
	UserID string
}

// ShortenURLBatchResponse represents a response with multiple shortened URLs.
type ShortenURLBatchResponse struct {
	URLs []BatchResponse
}

// ShortenURLBatch creates multiple shortened URLs at once.
func (s *Service) ShortenURLBatch(ctx context.Context, req ShortenURLBatchRequest) (*ShortenURLBatchResponse, error) {
	if len(req.URLs) == 0 {
		return nil, fmt.Errorf("empty batch request")
	}

	batchResponses := make([]BatchResponse, 0, len(req.URLs))
	urlsToSave := make(map[string]string, len(req.URLs))

	for _, urlReq := range req.URLs {
		shortURL := generateShortURL()
		err := s.storage.AddURL(shortURL, urlReq.OriginalURL, req.UserID)
		
		if err != nil && err.Error() != "URL already exists" {
			return nil, fmt.Errorf("failed to save URL mapping: %w", err)
		}
		
		batchResponses = append(batchResponses, BatchResponse{
			CorrelationID: urlReq.CorrelationID,
			ShortURL:      fmt.Sprintf("%s/%s", s.config.BaseURL, shortURL),
		})
		urlsToSave[shortURL] = urlReq.OriginalURL
	}

	// Save to file if configured
	if s.config.FileStorage != "" && len(urlsToSave) > 0 {
		if err := storage.SaveURLMappings(s.config.FileStorage, urlsToSave); err != nil {
			log.Printf("Warning: Failed to save URL mappings to file: %v", err)
		}
	}

	return &ShortenURLBatchResponse{
		URLs: batchResponses,
	}, nil
}

// GetUserURLsRequest represents a request to get user's URLs.
type GetUserURLsRequest struct {
	UserID string
}

// UserURL represents a URL created by a user.
type UserURL struct {
	ShortURL    string
	OriginalURL string
}

// GetUserURLsResponse represents a response with user's URLs.
type GetUserURLsResponse struct {
	URLs []UserURL
}

// GetUserURLs retrieves all URLs created by the authenticated user.
func (s *Service) GetUserURLs(ctx context.Context, req GetUserURLsRequest) (*GetUserURLsResponse, error) {
	if req.UserID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	urls, err := s.storage.GetURLsByUser(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user URLs: %w", err)
	}

	response := make([]UserURL, 0, len(urls))
	for short, original := range urls {
		response = append(response, UserURL{
			ShortURL:    fmt.Sprintf("%s/%s", s.config.BaseURL, short),
			OriginalURL: original,
		})
	}

	return &GetUserURLsResponse{
		URLs: response,
	}, nil
}

// DeleteUserURLsRequest represents a request to delete user's URLs.
type DeleteUserURLsRequest struct {
	ShortURLs []string
	UserID    string
}

// DeleteUserURLsResponse represents a response for URL deletion.
type DeleteUserURLsResponse struct {
	Success bool
}

// DeleteUserURLs marks URLs for deletion.
func (s *Service) DeleteUserURLs(ctx context.Context, req DeleteUserURLsRequest) (*DeleteUserURLsResponse, error) {
	err := s.storage.DeleteURLs(req.ShortURLs, req.UserID)
	if err != nil {
		return &DeleteUserURLsResponse{Success: false}, fmt.Errorf("failed to delete URLs: %w", err)
	}

	return &DeleteUserURLsResponse{Success: true}, nil
}

// PingRequest represents a health check request.
type PingRequest struct{}

// PingResponse represents a health check response.
type PingResponse struct {
	OK bool
}

// Ping performs a health check on the storage.
func (s *Service) Ping(ctx context.Context, req PingRequest) (*PingResponse, error) {
	err := s.storage.Ping()
	return &PingResponse{OK: err == nil}, err
}

// GetStatsRequest represents a request for service statistics.
type GetStatsRequest struct{}

// GetStatsResponse represents a response with service statistics.
type GetStatsResponse struct {
	URLs  int
	Users int
}

// GetStats returns service statistics.
func (s *Service) GetStats(ctx context.Context, req GetStatsRequest) (*GetStatsResponse, error) {
	urlCount, userCount, err := s.storage.GetStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	return &GetStatsResponse{
		URLs:  urlCount,
		Users: userCount,
	}, nil
}

// generateShortURL generates a random short URL identifier.
func generateShortURL() string {
	b := make([]byte, 6)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	return base64.URLEncoding.EncodeToString(b)[:6]
} 