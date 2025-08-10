package service

import (
	"errors"
	"math/rand"
	"sync"
	"time"

	"github.com/noedaka/go-url-shortener/internal/storage"
)

const shortIDLength = 6
const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type URLStorer interface {
	GetURL(shortID string) (string, error)
	ShortenURL(originalURL string) (string, error)
}

type urlStorage struct {
	mu      sync.RWMutex
	urls    map[string]string
	storage storage.URLStorage
}

func NewURLStorage(storage storage.URLStorage) URLStorer {
	s := &urlStorage{
		urls:    make(map[string]string),
		storage: storage,
	}

	if data, err := storage.Load(); err == nil {
		s.urls = data
	}

	return s
}

func (s *urlStorage) GetURL(shortID string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	url, exists := s.urls[shortID]
	if !exists {
		return "", errors.New("URL not found")
	}

	return url, nil
}

func (s *urlStorage) ShortenURL(originalURL string) (string, error) {
	var shortID string
	attempts := 0

	for {
		shortID = generateShortID()

		if s.isShortIDUnique(shortID) {
			break
		}

		attempts++
		if attempts > 10 {
			return "", errors.New("failed to generate unique ID after multiple attempts")
		}
	}

	err := s.saveShortID(shortID, originalURL)
	return shortID, err
}

func (s *urlStorage) isShortIDUnique(shortID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.urls[shortID]
	return !exists
}

func (s *urlStorage) saveShortID(shortID, originalURL string) error {
	s.mu.Lock()
	s.urls[shortID] = originalURL
	s.mu.Unlock()

	return s.storage.Save(shortID, originalURL)
}

func generateShortID() string {
	globalRand := rand.New(rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, shortIDLength)
	for i := range b {
		b[i] = charset[globalRand.Intn(len(charset))]
	}

	return string(b)
}
