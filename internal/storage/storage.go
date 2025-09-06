package storage

import (
	"context"

	"github.com/noedaka/go-url-shortener/internal/model"
)

type URLStorage interface {
	Save(shortURL, originalURL, userID string) error
	Get(shortURL string) (string, error)
	GetByUser(userID string) ([]model.URLPair, error)
	DeleteByUser(ctx context.Context, userID string, shortURL []string) error
}
