package storage

type Storage interface {
	AddURL(shortURL, originalURL string) error
	AddURLs(urls map[string]string) error
	GetURL(shortURL string) (string, bool)
	GetAllURLs() map[string]string
	GetShortURLByOriginalURL(originalURL string) (string, bool)
	Ping() error
	Close() error
}
