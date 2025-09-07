package storage

import (
	"context"

	"github.com/noedaka/go-url-shortener/internal/model"
)

type URLStorage interface {
	Save(ctx context.Context, shortURL, originalURL, userID string) error
	Get(ctx context.Context, shortURL string) (string, error)
	GetByUser(ctx context.Context, userID string) ([]model.URLPair, error)
	DeleteByUser(ctx context.Context, userID string, shortURL []string) error
}
