// Модуль service определяет операции над URL.
package service

import (
	"context"
	"math/rand"
	"time"

	"github.com/noedaka/go-url-shortener/internal/model"
	"github.com/noedaka/go-url-shortener/internal/storage"
)

// ShortenerService реализует операции над URL.
type ShortenerService struct {
	// storage для работы с хранилищем URL.
	storage storage.URLStorage
	// BaseURL представляет адрес используемый сервером приложения.
	BaseURL string
	rand    *rand.Rand
}

// NewShortenerService создает новый экземпляр ShortenerService.
func NewShortenerService(storage storage.URLStorage, baseURL string) *ShortenerService {
	return &ShortenerService{
		storage: storage,
		BaseURL: baseURL,
		rand:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GetURL возвращает полный URL по его сокращенному ID.
func (s *ShortenerService) GetURL(ctx context.Context, shortID string) (string, error) {
	return s.storage.Get(ctx, shortID)
}

// GetURLByUser возращает все пары сокращенного URL и оригинального URL, когда либо сокращенные указанным пользователем.
func (s *ShortenerService) GetURLByUser(ctx context.Context, userID string) ([]model.URLPair, error) {
	urlPairs, err := s.storage.GetByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	for i := range urlPairs {
		urlPairs[i].ShortURL = s.BaseURL + "/" + urlPairs[i].ShortURL
	}
	return urlPairs, nil
}

// DeleteShortURLSByUser удаляет сокращенные URL указанного пользователя.
func (s *ShortenerService) DeleteShortURLSByUser(ctx context.Context, userID string, shortURL []string) error {
	if len(shortURL) == 0 {
		return nil
	}
	if err := s.storage.DeleteByUser(ctx, userID, shortURL); err != nil {
		return err
	}

	return nil
}

// ShortenURL создает сокращенный URL и сохраняет его в хранилище указанного пользователя.
func (s *ShortenerService) ShortenURL(ctx context.Context, originalURL, userID string) (string, error) {
	shortID := s.generateShortID()
	err := s.storage.Save(ctx, shortID, originalURL, userID)

	if err != nil {
		return "", err
	}

	return shortID, nil
}

// ShortenMultipleURLS создает сокращенные URL для слайса URL.
func (s *ShortenerService) ShortenMultipleURLS(ctx context.Context, batchRequest []model.BatchRequest, userID string) ([]model.BatchResponse, error) {
	var batchResponse []model.BatchResponse
	for _, request := range batchRequest {
		shortURL, err := s.ShortenURL(ctx, request.URL, userID)
		if err != nil {
			return nil, err
		}

		response := model.BatchResponse{
			CorrelationID: request.CorrelationID,
			ShortURL:      s.BaseURL + "/" + shortURL,
		}

		batchResponse = append(batchResponse, response)
	}

	return batchResponse, nil
}

func (s *ShortenerService) GetStats(ctx context.Context) (*model.Stats, error) {
	return s.storage.GetStats(ctx)
}

func (s *ShortenerService) generateShortID() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[s.rand.Intn(len(charset))]
	}
	return string(b)
}
