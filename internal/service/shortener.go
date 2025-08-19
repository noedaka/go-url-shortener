// service/service.go
package service

import (
	"math/rand"
	"time"

	"github.com/noedaka/go-url-shortener/internal/storage"
)

type ShortenerService struct {
	storage storage.URLStorage
	rand    *rand.Rand
}

func NewShortenerService(storage storage.URLStorage) *ShortenerService {
	return &ShortenerService{
		storage: storage,
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

func (s *ShortenerService) generateShortID() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[s.rand.Intn(len(charset))]
	}
	return string(b)
}