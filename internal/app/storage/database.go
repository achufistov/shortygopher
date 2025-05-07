package storage

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/achufistov/shortygopher.git/internal/app/models"
	_ "github.com/lib/pq"
)

type DBStorage struct {
	db *sql.DB
}

func NewDBStorage(dsn string) (*DBStorage, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	// Create tables if they don't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS urls (
			short_url VARCHAR(255) PRIMARY KEY,
			original_url TEXT NOT NULL,
			user_id VARCHAR(255) NOT NULL,
			deleted BOOLEAN DEFAULT FALSE
		)
	`)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create tables: %v", err)
	}

	return &DBStorage{db: db}, nil
}

func (s *DBStorage) AddURL(shortURL, originalURL, userID string) error {
	_, err := s.db.Exec(
		"INSERT INTO urls (short_url, original_url, user_id) VALUES ($1, $2, $3) ON CONFLICT (short_url) DO UPDATE SET original_url = $2, user_id = $3",
		shortURL, originalURL, userID,
	)
	return err
}

func (s *DBStorage) AddURLs(urls map[string]string, userID string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO urls (short_url, original_url, user_id) 
		VALUES ($1, $2, $3) 
		ON CONFLICT (short_url) DO UPDATE 
		SET original_url = $2, user_id = $3
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for shortURL, originalURL := range urls {
		_, err = stmt.Exec(shortURL, originalURL, userID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *DBStorage) GetURL(shortURL string) (string, error) {
	var originalURL string
	var deleted bool
	err := s.db.QueryRow(
		"SELECT original_url, deleted FROM urls WHERE short_url = $1",
		shortURL,
	).Scan(&originalURL, &deleted)

	if err == sql.ErrNoRows {
		return "", errors.New("URL not found")
	}
	if err != nil {
		return "", err
	}
	if deleted {
		return "", errors.New("URL is deleted")
	}
	return originalURL, nil
}

func (s *DBStorage) GetURLsByUser(userID string) (map[string]string, error) {
	rows, err := s.db.Query(
		"SELECT short_url, original_url FROM urls WHERE user_id = $1 AND deleted = FALSE",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	urls := make(map[string]string)
	for rows.Next() {
		var shortURL, originalURL string
		if err := rows.Scan(&shortURL, &originalURL); err != nil {
			return nil, err
		}
		urls[shortURL] = originalURL
	}
	return urls, rows.Err()
}

func (s *DBStorage) GetAllURLs() map[string]string {
	rows, err := s.db.Query("SELECT short_url, original_url FROM urls WHERE deleted = FALSE")
	if err != nil {
		return make(map[string]string)
	}
	defer rows.Close()

	urls := make(map[string]string)
	for rows.Next() {
		var shortURL, originalURL string
		if err := rows.Scan(&shortURL, &originalURL); err != nil {
			return make(map[string]string)
		}
		urls[shortURL] = originalURL
	}
	return urls
}

func (s *DBStorage) GetShortURLByOriginalURL(originalURL string) (string, bool) {
	var shortURL string
	err := s.db.QueryRow(
		"SELECT short_url FROM urls WHERE original_url = $1 AND deleted = FALSE",
		originalURL,
	).Scan(&shortURL)

	if err == sql.ErrNoRows {
		return "", false
	}
	return shortURL, true
}

func (s *DBStorage) GetUserURLs(userID string) ([]models.URL, error) {
	rows, err := s.db.Query(
		"SELECT short_url, original_url FROM urls WHERE user_id = $1 AND deleted = FALSE",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var urls []models.URL
	for rows.Next() {
		var url models.URL
		if err := rows.Scan(&url.ShortURL, &url.OriginalURL); err != nil {
			return nil, err
		}
		url.UserID = userID
		urls = append(urls, url)
	}
	return urls, rows.Err()
}

func (s *DBStorage) DeleteURLs(shortURLs []string, userID string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		UPDATE urls 
		SET deleted = TRUE 
		WHERE short_url = $1 AND user_id = $2
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, shortURL := range shortURLs {
		_, err = stmt.Exec(shortURL, userID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *DBStorage) Ping() error {
	return s.db.Ping()
}

func (s *DBStorage) Close() error {
	return s.db.Close()
}
