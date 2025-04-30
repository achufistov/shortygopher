package storage

type Storage interface {
	AddURL(shortURL, originalURL, userID string) error
	AddURLs(urls map[string]string, userID string) error
	GetURL(shortURL string) (string, bool)
	GetURLsByUser(userID string) (map[string]string, error)
	GetAllURLs() map[string]string
	GetShortURLByOriginalURL(originalURL string) (string, bool)
	Ping() error
	Close() error
}
