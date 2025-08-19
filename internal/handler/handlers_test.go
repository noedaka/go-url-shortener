package handler

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/noedaka/go-url-shortener/internal/service"
	"github.com/stretchr/testify/assert"
)

type MockStorage struct {
	urls map[string]string
	err  error
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		urls: make(map[string]string),
	}
}

func (m *MockStorage) Save(shortURL, originalURL string) error {
	if m.err != nil {
		return m.err
	}
	m.urls[shortURL] = originalURL
	return nil
}

func (m *MockStorage) Get(shortURL string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	url, exists := m.urls[shortURL]
	if !exists {
		return "", errors.New("URL not found")
	}
	return url, nil
}

func (m *MockStorage) Close() error {
	return nil
}

func (m *MockStorage) SetError(err error) {
	m.err = err
}

func (m *MockStorage) AddURL(shortID, originalURL string) {
	m.urls[shortID] = originalURL
}

func TestHandler_ShortenURLHandler(t *testing.T) {
	mockStorage := NewMockStorage()
	svc := service.NewShortenerService(mockStorage)
	h := NewHandler(*svc, "http://localhost:8080", nil)

	r := chi.NewRouter()
	r.Post("/", h.ShortenURLHandler)

	type want struct {
		statusCode int
		contains   string
	}

	tests := []struct {
		name   string
		method string
		body   string
		want   want
	}{
		{
			name:   "POST valid",
			method: http.MethodPost,
			body:   "https://example.com",
			want: want{
				statusCode: http.StatusCreated,
				contains:   "http://localhost:8080/",
			},
		},
		{
			name:   "POST empty body",
			method: http.MethodPost,
			body:   "",
			want: want{
				statusCode: http.StatusCreated,
			},
		},
		{
			name:   "Wrong method",
			method: http.MethodPut,
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "text/plain")
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.statusCode, rr.Code)
			if tt.want.contains != "" {
				assert.Contains(t, rr.Body.String(), tt.want.contains)
			}
		})
	}
}

func TestHandler_ShortIdHandler(t *testing.T) {
	mockStorage := NewMockStorage()
	svc := service.NewShortenerService(mockStorage)
	h := NewHandler(*svc, "http://localhost:8080", nil)

	r := chi.NewRouter()
	r.Get("/{id}", h.ShortIDHandler)

	type want struct {
		statusCode int
		location   string
	}

	tests := []struct {
		name    string
		method  string
		id      string
		prepare func()
		want    want
	}{
		{
			name:   "GET valid",
			method: http.MethodGet,
			id:     "validID",
			prepare: func() {
				mockStorage.AddURL("validID", "https://example.com")
			},
			want: want{
				statusCode: http.StatusTemporaryRedirect,
				location:   "https://example.com",
			},
		},
		{
			name:   "GET nonexistent id",
			method: http.MethodGet,
			id:     "invalidID",
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage.err = nil

			if tt.prepare != nil {
				tt.prepare()
			}

			req := httptest.NewRequest(tt.method, "/"+tt.id, nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.statusCode, rr.Code)
			if tt.want.location != "" {
				assert.Equal(t, tt.want.location, rr.Header().Get("Location"))
			}
		})
	}
}

func TestHandler_APIShortenerHandler(t *testing.T) {
	mockStorage := NewMockStorage()
	svc := service.NewShortenerService(mockStorage)
	h := NewHandler(*svc, "http://localhost:8080", nil)

	r := chi.NewRouter()
	r.Post("/api/shorten", h.APIShortenerHandler)

	type want struct {
		statusCode  int
		contentType string
		contains    string
	}

	tests := []struct {
		name    string
		method  string
		body    string
		prepare func(s *MockStorage)
		want    want
	}{
		{
			name:   "POST valid JSON",
			method: http.MethodPost,
			body:   `{"url": "https://example.com"}`,
			want: want{
				statusCode:  http.StatusCreated,
				contentType: "application/json",
				contains:    `"result":"http://localhost:8080/`,
			},
		},
		{
			name:   "POST invalid JSON",
			method: http.MethodPost,
			body:   `invalid json`,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:   "Wrong method",
			method: http.MethodPut,
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
		},
		{
			name:    "Service error",
			method:  http.MethodPost,
			body:    `{"url": "https://example.com"}`,
			prepare: func(s *MockStorage) { s.SetError(errors.New("service error")) },
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.prepare != nil {
				tt.prepare(mockStorage)
			}

			req := httptest.NewRequest(tt.method, "/api/shorten", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.statusCode, rr.Code)

			if tt.want.contentType != "" {
				assert.Equal(t, tt.want.contentType, rr.Header().Get("Content-type"))
			}

			if tt.want.contains != "" {
				assert.Contains(t, rr.Body.String(), tt.want.contains)
			}
		})
	}
}
