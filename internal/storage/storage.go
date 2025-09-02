package storage

import "github.com/noedaka/go-url-shortener/internal/model"

type URLStorage interface {
	Save(shortURL, originalURL, userID string) error
	Get(shortURL string) (string, error)
	GetByUser(userID string) ([]model.URLPair, error)
}
