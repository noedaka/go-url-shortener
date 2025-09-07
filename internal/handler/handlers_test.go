package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/noedaka/go-url-shortener/internal/config"
	"github.com/noedaka/go-url-shortener/internal/model"
	"github.com/noedaka/go-url-shortener/internal/service"
	"github.com/stretchr/testify/assert"
)

type MockStorage struct {
	urls        map[string]string
	users       map[string]map[string]string
	err         error
	deletedArgs []struct {
		userID    string
		shortURLs []string
	}
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		urls:  make(map[string]string),
		users: make(map[string]map[string]string),
	}
}

func (m *MockStorage) Save(ctx context.Context, shortURL, originalURL, userID string) error {
	if m.err != nil {
		return m.err
	}
	m.urls[shortURL] = originalURL

	if _, exists := m.users[userID]; !exists {
		m.users[userID] = make(map[string]string)
	}
	m.users[userID][shortURL] = originalURL

	return nil
}

func (m *MockStorage) Get(ctx context.Context, shortURL string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	url, exists := m.urls[shortURL]
	if !exists {
		return "", errors.New("URL not found")
	}
	return url, nil
}

func (m *MockStorage) GetByUser(ctx context.Context, userID string) ([]model.URLPair, error) {
	if m.err != nil {
		return nil, m.err
	}

	userURLs, exists := m.users[userID]
	if !exists {
		return []model.URLPair{}, nil
	}

	var pairs []model.URLPair
	for shortURL, originalURL := range userURLs {
		pairs = append(pairs, model.URLPair{
			ShortURL:    shortURL,
			OriginalURL: originalURL,
		})
	}

	return pairs, nil
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

func (m *MockStorage) AddURLForUser(shortID, originalURL, userID string) {
	m.urls[shortID] = originalURL
	if _, exists := m.users[userID]; !exists {
		m.users[userID] = make(map[string]string)
	}
	m.users[userID][shortID] = originalURL
}

func (m *MockStorage) DeleteByUser(ctx context.Context, userID string, shortURLs []string) error {
	m.deletedArgs = append(m.deletedArgs, struct {
		userID    string
		shortURLs []string
	}{userID: userID, shortURLs: shortURLs})

	if m.err != nil {
		return m.err
	}

	if userURLs, exists := m.users[userID]; exists {
		for _, shortURL := range shortURLs {
			delete(userURLs, shortURL)
		}
	}
	return nil
}

func (m *MockStorage) ResetDeletedArgs() {
	m.deletedArgs = nil
}

func (m *MockStorage) GetDeletedArgs() []struct {
	userID    string
	shortURLs []string
} {
	return m.deletedArgs
}

func withUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, config.UserIDKey, userID)
}

func TestHandler_ShortenURLHandler(t *testing.T) {
	mockStorage := NewMockStorage()
	svc := service.NewShortenerService(mockStorage, "http://localhost:8080")
	h := NewHandler(*svc, nil)

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
		userID string
		want   want
	}{
		{
			name:   "POST valid",
			method: http.MethodPost,
			body:   "https://example.com",
			userID: "test-user",
			want: want{
				statusCode: http.StatusCreated,
				contains:   "http://localhost:8080/",
			},
		},
		{
			name:   "POST empty body",
			method: http.MethodPost,
			body:   "",
			userID: "test-user",
			want: want{
				statusCode: http.StatusCreated,
			},
		},
		{
			name:   "Wrong method",
			method: http.MethodPut,
			userID: "test-user",
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "text/plain")

			ctx := withUserID(req.Context(), tt.userID)
			req = req.WithContext(ctx)

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
	svc := service.NewShortenerService(mockStorage, "http://localhost:8080")
	h := NewHandler(*svc, nil)

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
	svc := service.NewShortenerService(mockStorage, "http://localhost:8080")
	h := NewHandler(*svc, nil)

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
		userID  string
		prepare func(s *MockStorage)
		want    want
	}{
		{
			name:   "POST valid JSON",
			method: http.MethodPost,
			body:   `{"url": "https://example.com"}`,
			userID: "test-user",
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
			userID: "test-user",
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:   "Wrong method",
			method: http.MethodPut,
			userID: "test-user",
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
		},
		{
			name:    "Service error",
			method:  http.MethodPost,
			body:    `{"url": "https://example.com"}`,
			userID:  "test-user",
			prepare: func(s *MockStorage) { s.SetError(errors.New("service error")) },
			want: want{
				statusCode: http.StatusInternalServerError,
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

			ctx := withUserID(req.Context(), tt.userID)
			req = req.WithContext(ctx)

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

func TestHandler_ShortenBatchHandler(t *testing.T) {
	mockStorage := NewMockStorage()
	svc := service.NewShortenerService(mockStorage, "http://localhost:8080")
	h := NewHandler(*svc, nil)

	r := chi.NewRouter()
	r.Post("/api/shorten/batch", h.ShortenBatchHandler)

	type want struct {
		statusCode          int
		contentType         string
		contains            string
		checkCorrelationIDs bool
		expectedIDs         []string
	}

	tests := []struct {
		name   string
		method string
		body   string
		userID string
		want   want
	}{
		{
			name:   "POST valid batch",
			method: http.MethodPost,
			body:   `[{"correlation_id": "test1", "original_url": "https://example.com/1"}, {"correlation_id": "test2", "original_url": "https://example.com/2"}]`,
			userID: "test-user",
			want: want{
				statusCode:          http.StatusCreated,
				contentType:         "application/json",
				checkCorrelationIDs: true,
				expectedIDs:         []string{"test1", "test2"},
			},
		},
		{
			name:   "POST empty batch",
			method: http.MethodPost,
			body:   `[]`,
			userID: "test-user",
			want: want{
				statusCode: http.StatusBadRequest,
				contains:   "empty JSON body",
			},
		},
		{
			name:   "POST invalid JSON",
			method: http.MethodPost,
			body:   `invalid json`,
			userID: "test-user",
			want: want{
				statusCode: http.StatusBadRequest,
				contains:   "cannot decode request JSON body",
			},
		},
		{
			name:   "Wrong method",
			method: http.MethodGet,
			body:   `[{"correlation_id": "test1", "original_url": "https://example.com/1"}]`,
			userID: "test-user",
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/shorten/batch", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")

			ctx := withUserID(req.Context(), tt.userID)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.want.statusCode, rr.Code)

			if tt.want.contains != "" {
				assert.Contains(t, rr.Body.String(), tt.want.contains)
			}

			if tt.want.contentType != "" {
				assert.Equal(t, tt.want.contentType, rr.Header().Get("Content-Type"))
			}

			if tt.want.checkCorrelationIDs {
				var responses []model.BatchResponse
				err := json.Unmarshal(rr.Body.Bytes(), &responses)
				assert.NoError(t, err)

				assert.Equal(t, len(tt.want.expectedIDs), len(responses))

				for i, expectedID := range tt.want.expectedIDs {
					assert.Equal(t, expectedID, responses[i].CorrelationID)
					assert.Contains(t, responses[i].ShortURL, "http://localhost:8080/")
				}
			}
		})
	}
}

func TestHandler_APIUserUrlsHandler(t *testing.T) {
	mockStorage := NewMockStorage()
	svc := service.NewShortenerService(mockStorage, "http://localhost:8080")
	h := NewHandler(*svc, nil)

	r := chi.NewRouter()
	r.Get("/api/user/urls", h.APIUserUrlsHandler)

	tests := []struct {
		name       string
		method     string
		userID     string
		prepare    func(s *MockStorage)
		wantStatus int
		wantBody   string
	}{
		{
			name:   "Success with URLs",
			method: http.MethodGet,
			userID: "test-user",
			prepare: func(s *MockStorage) {
				s.AddURLForUser("abc123", "https://example.com/1", "test-user")
				s.AddURLForUser("def456", "https://example.com/2", "test-user")
			},
			wantStatus: http.StatusOK,
			wantBody:   `[{"short_url":"http://localhost:8080/abc123","original_url":"https://example.com/1"},{"short_url":"http://localhost:8080/def456","original_url":"https://example.com/2"}]`,
		},
		{
			name:       "Unauthorized",
			method:     http.MethodGet,
			userID:     "",
			prepare:    func(s *MockStorage) {},
			wantStatus: http.StatusUnauthorized,
			wantBody:   "unauthorized",
		},
		{
			name:   "Service error",
			method: http.MethodGet,
			userID: "test-user",
			prepare: func(s *MockStorage) {
				s.SetError(errors.New("service error"))
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   "cannot get urls by user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage.SetError(nil)

			if tt.prepare != nil {
				tt.prepare(mockStorage)
			}

			req := httptest.NewRequest(tt.method, "/api/user/urls", nil)

			if tt.userID != "" {
				ctx := withUserID(req.Context(), tt.userID)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)

			if tt.wantBody != "" {
				assert.Contains(t, rr.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestHandler_APIDeleteShortURLSHandler(t *testing.T) {
	mockStorage := NewMockStorage()
	svc := service.NewShortenerService(mockStorage, "http://localhost:8080")
	h := NewHandler(*svc, nil)

	r := chi.NewRouter()
	r.Delete("/api/user/urls", h.APIDeleteShortURLSHandler)

	tests := []struct {
		name       string
		method     string
		body       string
		userID     string
		prepare    func(s *MockStorage)
		wantStatus int
	}{
		{
			name:   "Success",
			method: http.MethodDelete,
			body:   `["short1", "short2"]`,
			userID: "test-user",
			prepare: func(s *MockStorage) {
				s.AddURLForUser("short1", "https://example.com/1", "test-user")
				s.AddURLForUser("short2", "https://example.com/2", "test-user")
			},
			wantStatus: http.StatusAccepted,
		},
		{
			name:       "Unauthorized",
			method:     http.MethodDelete,
			body:       `["short1", "short2"]`,
			userID:     "",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "Bad JSON",
			method:     http.MethodDelete,
			body:       `invalid json`,
			userID:     "test-user",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "Service error",
			method: http.MethodDelete,
			body:   `["short1", "short2"]`,
			userID: "test-user",
			prepare: func(s *MockStorage) {
				s.SetError(errors.New("service error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage.ResetDeletedArgs()
			mockStorage.SetError(nil)

			if tt.prepare != nil {
				tt.prepare(mockStorage)
			}

			req := httptest.NewRequest(tt.method, "/api/user/urls", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")

			if tt.userID != "" {
				ctx := withUserID(req.Context(), tt.userID)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
		})
	}
}
