package service

import (
	"testing"
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

func (m *MockStorage) Save(shortURL, originalURL string) error {
	m.data[shortURL] = originalURL
	return nil
}

func (m *MockStorage) Get(shortURL string) (string, error) {
	if url, exists := m.data[shortURL]; exists {
		return url, nil
	}
	return "", &URLNotFoundError{ShortURL: shortURL}
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
			mockStorage := NewMockStorage()
			service := NewShortenerService(mockStorage, "")

			shortID, err := service.ShortenURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ShortenURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(shortID) != 6 { // shortIDLength не определен, используем 6
				t.Errorf("ShortenURL() got len = %d, want %d", len(shortID), 6)
			}

			if !tt.wantErr {
				gotURL, err := service.GetURL(shortID)
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
	mockStorage := NewMockStorage()
	service := NewShortenerService(mockStorage, "")
	_, err := service.GetURL("nonexistent")
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
