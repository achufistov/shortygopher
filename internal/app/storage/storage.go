package storage

type Storage interface {
	AddURL(shortURL, originalURL, userID string) error
	AddURLs(urls map[string]string) error
	GetURL(shortURL string) (string, bool)
	GetAllURLs() map[string]string
	GetShortURLByOriginalURL(originalURL string) (string, bool)
	GetUserURLs(userID string) (map[string]string, error)
	Ping() error
	Close() error
}
