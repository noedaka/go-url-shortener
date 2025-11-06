package service

import (
	"context"
	"testing"

	"github.com/noedaka/go-url-shortener/internal/model"
)

type MockStorage struct {
	data    map[string]string
	baseURL string
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		data:    make(map[string]string),
		baseURL: "",
	}
}

func (m *MockStorage) Save(ctx context.Context, shortURL, originalURL, userID string) error {
	m.data[shortURL] = originalURL
	return nil
}

func (m *MockStorage) Get(ctx context.Context, shortURL string) (string, error) {
	if url, exists := m.data[shortURL]; exists {
		return url, nil
	}
	return "", &URLNotFoundError{ShortURL: shortURL}
}

func (m *MockStorage) GetByUser(ctx context.Context, userID string) ([]model.URLPair, error) {
	return nil, nil
}

func (m *MockStorage) DeleteByUser(ctx context.Context, userID string, shortURL []string) error {
	return nil
}

func TestShortenerService(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"Valid URL", "https://example.com", false},
		{"Empty URL", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockStorage := NewMockStorage()
			service := NewShortenerService(mockStorage, "")

			shortID, err := service.ShortenURL(ctx, tt.url, "")
			if (err != nil) != tt.wantErr {
				t.Errorf("ShortenURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(shortID) != 6 {
				t.Errorf("ShortenURL() got len = %d, want %d", len(shortID), 6)
			}

			if !tt.wantErr {
				gotURL, err := service.GetURL(ctx, shortID)
				if err != nil {
					t.Errorf("GetURL() error = %v", err)
				}
				if gotURL != tt.url {
					t.Errorf("GetURL() got = %v, want %v", gotURL, tt.url)
				}
			}
		})
	}
}

func TestGetURL_NotFound(t *testing.T) {
	ctx := context.Background()
	mockStorage := NewMockStorage()
	service := NewShortenerService(mockStorage, "")
	_, err := service.GetURL(ctx, "nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent URL")
	}
}

type URLNotFoundError struct {
	ShortURL string
}

func (e *URLNotFoundError) Error() string {
	return "URL not found: " + e.ShortURL
}

func BenchmarkShortenURL(b *testing.B) {
	ctx := context.Background()
	mockStorage := NewMockStorage()
	service := NewShortenerService(mockStorage, "http://localhost:8080")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.ShortenURL(ctx, "https://example.com/very/long/url/that/needs/shortening", "user123")
		if err != nil {
			b.Fatalf("ShortenURL failed: %v", err)
		}
	}
}

func BenchmarkGetURL(b *testing.B) {
	ctx := context.Background()
	mockStorage := NewMockStorage()
	service := NewShortenerService(mockStorage, "http://localhost:8080")

	shortID, err := service.ShortenURL(ctx, "https://example.com", "user123")
	if err != nil {
		b.Fatalf("Setup failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GetURL(ctx, shortID)
		if err != nil {
			b.Fatalf("GetURL failed: %v", err)
		}
	}
}

func BenchmarkGenerateShortID(b *testing.B) {
	mockStorage := NewMockStorage()
	service := NewShortenerService(mockStorage, "http://localhost:8080")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.ShortenURL(context.Background(), "https://example.com", "user123")
	}
}

func BenchmarkShortenMultipleURLS(b *testing.B) {
	ctx := context.Background()
	mockStorage := NewMockStorage()
	service := NewShortenerService(mockStorage, "http://localhost:8080")

	batchRequests := []model.BatchRequest{
		{CorrelationID: "1", URL: "https://example.com/1"},
		{CorrelationID: "2", URL: "https://example.com/2"},
		{CorrelationID: "3", URL: "https://example.com/3"},
		{CorrelationID: "4", URL: "https://example.com/4"},
		{CorrelationID: "5", URL: "https://example.com/5"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.ShortenMultipleURLS(ctx, batchRequests, "user123")
		if err != nil {
			b.Fatalf("ShortenMultipleURLS failed: %v", err)
		}
	}
}

func BenchmarkGetURLByUser(b *testing.B) {
	ctx := context.Background()

	mockStorage := NewFakeStorageWithUserData()
	service := NewShortenerService(mockStorage, "http://localhost:8080")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GetURLByUser(ctx, "user123")
		if err != nil {
			b.Fatalf("GetURLByUser failed: %v", err)
		}
	}
}

func BenchmarkDeleteShortURLSByUser(b *testing.B) {
	ctx := context.Background()
	mockStorage := NewFakeStorageWithUserData()
	service := NewShortenerService(mockStorage, "http://localhost:8080")

	shortURLs := []string{"abc123", "def456", "ghi789"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := service.DeleteShortURLSByUser(ctx, "user123", shortURLs)
		if err != nil {
			b.Fatalf("DeleteShortURLSByUser failed: %v", err)
		}
	}
}

type FakeStorageWithUserData struct {
	data     map[string]string
	userURLs map[string][]model.URLPair
	baseURL  string
}

func NewFakeStorageWithUserData() *FakeStorageWithUserData {
	return &FakeStorageWithUserData{
		data:     make(map[string]string),
		userURLs: make(map[string][]model.URLPair),
		baseURL:  "",
	}
}

func (m *FakeStorageWithUserData) Save(ctx context.Context, shortURL, originalURL, userID string) error {
	m.data[shortURL] = originalURL
	if userID != "" {
		m.userURLs[userID] = append(m.userURLs[userID], model.URLPair{
			ShortURL:    shortURL,
			OriginalURL: originalURL,
		})
	}
	return nil
}

func (m *FakeStorageWithUserData) Get(ctx context.Context, shortURL string) (string, error) {
	if url, exists := m.data[shortURL]; exists {
		return url, nil
	}
	return "", &URLNotFoundError{ShortURL: shortURL}
}

func (m *FakeStorageWithUserData) GetByUser(ctx context.Context, userID string) ([]model.URLPair, error) {
	if urls, exists := m.userURLs[userID]; exists {
		return urls, nil
	}
	return []model.URLPair{}, nil
}

func (m *FakeStorageWithUserData) DeleteByUser(ctx context.Context, userID string, shortURLs []string) error {
	if _, exists := m.userURLs[userID]; exists {
		newURLs := []model.URLPair{}
		for _, pair := range m.userURLs[userID] {
			shouldKeep := true
			for _, toDelete := range shortURLs {
				if pair.ShortURL == toDelete {
					shouldKeep = false
					break
				}
			}
			if shouldKeep {
				newURLs = append(newURLs, pair)
			}
		}
		m.userURLs[userID] = newURLs

		for _, shortURL := range shortURLs {
			delete(m.data, shortURL)
		}
	}
	return nil
}
