package service

import (
	"errors"
	"math/rand"
	"sync"
)

const shortIDLenght = 6
const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type URLStorage struct {
	mu   sync.RWMutex      // Для потокобезопасности мапы
	urls map[string]string // ключ - ID; значение - ориг URL
}

func NewURLStorage() *URLStorage {
	return &URLStorage{
		urls: make(map[string]string),
	}
}

func (s *URLStorage) GetURL(shortID string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	url, exists := s.urls[shortID]
	if !exists {
		return "", errors.New("URL not found")
	}

	return url, nil
}

func (s *URLStorage) ShortenURL(originalURL string) (string, error) {
	shortID := generateShortID()
	err := s.saveShortID(shortID, originalURL)
	return shortID, err
}

func (s *URLStorage) saveShortID(shortID, originalURL string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.urls[shortID] = originalURL

	return nil
}

func generateShortID() string {
	b := make([]byte, shortIDLenght)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}

	return string(b)
}
