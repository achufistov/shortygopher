package storage

type Storage interface {
	AddURL(shortURL, originalURL string)
	AddURLs(urls map[string]string) error
	GetURL(shortURL string) (string, bool)
	GetAllURLs() map[string]string
	Ping() error
	Close() error
}
