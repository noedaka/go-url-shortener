package service

import (
	"testing"
)

func TestURLStorage(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"Valid URL", "https://example.com", false},
		{"Empty URL", "", false}, 
	}

	storage := NewURLStorage()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shortID, err := storage.ShortenURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ShortenURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(shortID) != shortIDLenght {
				t.Errorf("ShortenURL() got len = %d, want %d", len(shortID), shortIDLenght)
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
	storage := NewURLStorage()
	_, err := storage.GetURL("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent URL")
	}
}

func TestGenerateShortID(t *testing.T) {
	id1 := generateShortID()
	id2 := generateShortID()

	if len(id1) != shortIDLenght {
		t.Errorf("Wrong length: got %d, want %d", len(id1), shortIDLenght)
	}

	if id1 == id2 {
		t.Error("Expected different IDs")
	}
}