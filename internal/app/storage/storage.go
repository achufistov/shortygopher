// internal/app/storage/storage.go
package storage

type Storage interface {
	AddURL(shortURL, originalURL string)
	GetURL(shortURL string) (string, bool)
	GetAllURLs() map[string]string
	Ping() error
	Close() error
}
