package storage

import (
	"database/sql"
	"fmt"

	"github.com/lib/pq"
)

// DBStorage implements the Storage interface using PostgreSQL database.
// Provides persistent storage for URL mappings with support for user associations and soft deletes.
type DBStorage struct {
	db *sql.DB
}

// NewDBStorage creates a new DBStorage instance connected to PostgreSQL.
// Establishes database connection, verifies connectivity, and creates required tables.
// Returns error if connection fails or table creation fails.
func NewDBStorage(dsn string) (*DBStorage, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to establish connection for the database : %v", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	createTableQuery := `
	CREATE TABLE IF NOT EXISTS urls (
		id SERIAL PRIMARY KEY,
		url TEXT NOT NULL UNIQUE,
		short_url TEXT NOT NULL UNIQUE,
		user_id TEXT NOT NULL,
		is_deleted BOOLEAN DEFAULT FALSE
	);
	`
	if _, err = db.Exec(createTableQuery); err != nil {
		return nil, fmt.Errorf("unable to create database: %v", err)
	}

	return &DBStorage{db: db}, nil
}

// AddURL adds a new URL mapping to the database.
// Uses ON CONFLICT to handle duplicate URLs gracefully.
// Returns error if URL already exists or database operation fails.
func (s *DBStorage) AddURL(shortURL, originalURL, userID string) error {
	query := `
    INSERT INTO urls (url, short_url, user_id)
    VALUES ($1, $2, $3)
    ON CONFLICT (url) DO NOTHING
    RETURNING short_url;
    `
	var existingShortURL string
	err := s.db.QueryRow(query, originalURL, shortURL, userID).Scan(&existingShortURL)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("URL already exists")
		}
		return fmt.Errorf("failed to add URL to database: %v", err)
	}
	return nil
}

// AddURLs adds multiple URL mappings in a single database transaction.
// Rolls back all changes if any URL fails to insert.
// More efficient than multiple individual AddURL calls.
func (s *DBStorage) AddURLs(urls map[string]string, userID string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	query := `INSERT INTO urls (url, short_url, user_id) VALUES ($1, $2, $3)`
	for shortURL, originalURL := range urls {
		_, err := tx.Exec(query, originalURL, shortURL, userID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to add URL to database: %v", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

// GetURL retrieves the original URL and deletion status for a short URL.
// Returns original URL, existence flag, and deletion status.
func (s *DBStorage) GetURL(shortURL string) (string, bool, bool) {
	var originalURL string
	var isDeleted bool
	query := `SELECT url, is_deleted FROM urls WHERE short_url = $1`
	err := s.db.QueryRow(query, shortURL).Scan(&originalURL, &isDeleted)
	if err != nil {
		return "", false, false
	}
	return originalURL, true, isDeleted
}

// GetAllURLs retrieves all URL mappings from the database.
// Returns a map of short URL to original URL for all stored mappings.
func (s *DBStorage) GetAllURLs() map[string]string {
	urlMap := make(map[string]string)
	query := `SELECT short_url, url FROM urls`
	rows, err := s.db.Query(query)
	if err != nil {
		fmt.Printf("Failed to get URLs from database: %v\n", err)
		return urlMap
	}
	defer rows.Close()

	for rows.Next() {
		var shortURL, originalURL string
		if err := rows.Scan(&shortURL, &originalURL); err != nil {
			fmt.Printf("Failed to scan row: %v\n", err)
			continue
		}
		urlMap[shortURL] = originalURL
	}
	if err := rows.Err(); err != nil {
		fmt.Printf("Failed to iterate over rows: %v\n", err)
	}
	return urlMap
}

// GetShortURLByOriginalURL finds the short URL for a given original URL.
// Returns short URL and found flag. Useful for checking existing mappings.
func (s *DBStorage) GetShortURLByOriginalURL(originalURL string) (string, bool) {
	var shortURL string
	query := `SELECT short_url FROM urls WHERE url = $1`
	err := s.db.QueryRow(query, originalURL).Scan(&shortURL)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", false
		}
		fmt.Printf("Failed to get short URL by original URL: %v", err)
		return "", false
	}
	return shortURL, true
}

// GetURLsByUser retrieves all URL mappings created by a specific user.
// Returns a map of short URL to original URL for the specified user.
func (s *DBStorage) GetURLsByUser(userID string) (map[string]string, error) {
	urlMap := make(map[string]string)
	query := `SELECT short_url, url FROM urls WHERE user_id = $1`
	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query URLs by user: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var shortURL, originalURL string
		if err := rows.Scan(&shortURL, &originalURL); err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
		urlMap[shortURL] = originalURL
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %v", err)
	}

	return urlMap, nil
}

// DeleteURLs soft-deletes URLs by setting is_deleted flag to true.
// Uses PostgreSQL array operations for efficient batch deletion.
func (s *DBStorage) DeleteURLs(shortURLs []string, userID string) error {
	query := `UPDATE urls SET is_deleted = TRUE WHERE short_url = ANY($1)`
	_, err := s.db.Exec(query, pq.Array(shortURLs))
	return err
}

// Ping checks database connectivity.
// Returns error if database is unreachable.
func (s *DBStorage) Ping() error {
	return s.db.Ping()
}

// Close closes the database connection.
// Should be called when storage is no longer needed.
func (s *DBStorage) Close() error {
	return s.db.Close()
}

// GetStats returns statistics about the storage.
// Returns the number of URLs and unique users.
func (s *DBStorage) GetStats() (int, int, error) {
	var urlCount int
	var userCount int

	// Count total URLs
	urlQuery := `SELECT COUNT(*) FROM urls`
	err := s.db.QueryRow(urlQuery).Scan(&urlCount)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to count URLs: %v", err)
	}

	// Count unique users
	userQuery := `SELECT COUNT(DISTINCT user_id) FROM urls`
	err = s.db.QueryRow(userQuery).Scan(&userCount)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to count users: %v", err)
	}

	return urlCount, userCount, nil
}
