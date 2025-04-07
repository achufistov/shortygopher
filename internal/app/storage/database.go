package storage

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
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
        url TEXT NOT NULL,
        short_url TEXT NOT NULL UNIQUE
    );
    `
	if _, err = db.Exec(createTableQuery); err != nil {
		return nil, fmt.Errorf("unable to create database: %v", err)
	}

	return &DBStorage{db: db}, nil
}

func (s *DBStorage) AddURL(shortURL, originalURL string) {
	query := `INSERT INTO urls (url, short_url) VALUES ($1, $2)`
	_, err := s.db.Exec(query, originalURL, shortURL)
	if err != nil {
		fmt.Printf("Failed to add URL to database: %v\n", err)
	}
}

func (s *DBStorage) GetURL(shortURL string) (string, bool) {
	var originalURL string
	query := `SELECT url FROM urls WHERE short_url = $1`
	err := s.db.QueryRow(query, shortURL).Scan(&originalURL)
	if err != nil {
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
	return urlMap
}

func (s *DBStorage) Ping() error {
	return s.db.Ping()
}

func (s *DBStorage) Close() error {
	return s.db.Close()
}
