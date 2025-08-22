package service

import (
	"math/rand"
	"time"

	"github.com/noedaka/go-url-shortener/internal/model"
	"github.com/noedaka/go-url-shortener/internal/storage"
)

type ShortenerService struct {
	storage storage.URLStorage
	BaseURL string
	rand    *rand.Rand
}

func NewShortenerService(storage storage.URLStorage, baseURL string) *ShortenerService {
	return &ShortenerService{
		storage: storage,
		BaseURL: baseURL,
		rand:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s *ShortenerService) GetURL(shortID string) (string, error) {
	return s.storage.Get(shortID)
}

func (s *ShortenerService) ShortenURL(originalURL string) (string, error) {
	shortID := s.generateShortID()
	err := s.storage.Save(shortID, originalURL)

	if err != nil {
		return "", err
	}

	return shortID, nil
}

func (s *ShortenerService) ShortenMultipleURLS(batchRequest []model.BatchRequest) ([]model.BatchResponse, error) {
	var batchResponse []model.BatchResponse
	for _, request := range batchRequest {
		shortURL, err := s.ShortenURL(request.URL)
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

func (s *ShortenerService) generateShortID() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[s.rand.Intn(len(charset))]
	}
	return string(b)
}
