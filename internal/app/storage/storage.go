// Package storage provides interfaces and implementations for storing URL mappings.
package storage

// Storage defines the interface for storing shortened URLs.
// All implementations should support both in-memory and persistent storage.
//
// Example usage:
//
//	storage := NewURLStorage()
//	err := storage.AddURL("abc123", "https://example.com", "user1")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	originalURL, exists, isDeleted := storage.GetURL("abc123")
//	if exists && !isDeleted {
//		fmt.Printf("URL: %s\n", originalURL)
//	}
type Storage interface {
	// AddURL adds a new URL mapping.
	// Returns an error if the URL already exists or if there's a storage error.
	AddURL(shortURL, originalURL, userID string) error

	// AddURLs adds multiple URL mappings at once (batch operation).
	AddURLs(urls map[string]string, userID string) error

	// GetURL returns the original URL by short URL.
	// The second parameter indicates whether the URL exists.
	// The third parameter indicates whether the URL was deleted.
	GetURL(shortURL string) (string, bool, bool)

	// GetURLsByUser returns all URL mappings for the specified user.
	GetURLsByUser(userID string) (map[string]string, error)

	// GetAllURLs returns all URL mappings.
	GetAllURLs() map[string]string

	// GetShortURLByOriginalURL finds a short URL by original URL.
	GetShortURLByOriginalURL(originalURL string) (string, bool)

	// DeleteURLs marks the specified URLs as deleted for the specified user.
	DeleteURLs(shortURLs []string, userID string) error

	// Ping checks storage availability.
	Ping() error

	// Close closes the storage connection.
	Close() error

	// GetStats returns statistics about the storage.
	// Returns the number of URLs and unique users.
	GetStats() (int, int, error)
}
