package storage

import (
	"database/sql"
	"fmt"

	"github.com/lib/pq"
)

type DBStorage struct {
	db *sql.DB
}

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

func (s *DBStorage) GetURL(shortURL string) (string, bool) {
	var originalURL string
	var isDeleted bool
	query := `SELECT url, is_deleted FROM urls WHERE short_url = $1`
	err := s.db.QueryRow(query, shortURL).Scan(&originalURL, &isDeleted)
	if err != nil {
		return "", false
	}
	if isDeleted {
		return "", false
	}
	return originalURL, true
}

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

func (s *DBStorage) DeleteURLs(shortURLs []string, userID string) error {
	query := `UPDATE urls SET is_deleted = TRUE WHERE short_url = ANY($1)`
	_, err := s.db.Exec(query, pq.Array(shortURLs))
	return err
}

func (s *DBStorage) Ping() error {
	return s.db.Ping()
}

func (s *DBStorage) Close() error {
	return s.db.Close()
}
