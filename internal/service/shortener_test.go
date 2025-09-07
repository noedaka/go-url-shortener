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
