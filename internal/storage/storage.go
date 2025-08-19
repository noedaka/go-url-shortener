package storage

type URLStorage interface {
	Save(shortURL, originalURL string) error
	Get(shortURL string) (string, error)
}
