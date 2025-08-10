package service

import (
	"testing"
)

type MockStorage struct {
	data map[string]string
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		data: make(map[string]string),
	}
}

func (m *MockStorage) Save(shortURL, originalURL string) error {
	m.data[shortURL] = originalURL
	return nil
}

func (m *MockStorage) Load() (map[string]string, error) {
	return m.data, nil
}

func Test_URLStorage(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"Valid URL", "https://example.com", false},
		{"Empty URL", "", true},
	}

	mockStorage := NewMockStorage()
	storage := NewURLStorage(mockStorage)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shortID, err := storage.ShortenURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ShortenURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(shortID) != shortIDLength {
				t.Errorf("ShortenURL() got len = %d, want %d", len(shortID), shortIDLength)
			}

			if !tt.wantErr {
				gotURL, err := storage.GetURL(shortID)
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
	storage := NewURLStorage(mockStorage)
	_, err := storage.GetURL("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent URL")
	}
}
