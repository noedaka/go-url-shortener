// Модуль storage определяет интерфейс хранилища и его реализацию с in-memory хранением
// и хранением в PostgreSQL
package storage

import (
	"context"

	"github.com/noedaka/go-url-shortener/internal/model"
)

// URLStorage определяет интерфейс для работы с хранилищем данных
type URLStorage interface {
	// Save сохраняет сокращенный URL и оригинальный URL в хранилище указанного пользователя
	Save(ctx context.Context, shortURL, originalURL, userID string) error
	// Get возращает оригинальный URL по сокращенному
	Get(ctx context.Context, shortURL string) (string, error)
	// GetByUser возвращает все пары URL когда либо сокращенных указанным пользователем
	GetByUser(ctx context.Context, userID string) ([]model.URLPair, error)
	// DeleteByUser удаляет сокращенные URL указанного пользователя
	DeleteByUser(ctx context.Context, userID string, shortURL []string) error
	// GetStats возвращает количество сокращенных юрлов и количество пользователей
	GetStats(ctx context.Context) (*model.Stats, error)
}
