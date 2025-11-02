package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
	"github.com/noedaka/go-url-shortener/internal/config"
	"github.com/noedaka/go-url-shortener/internal/model"
	"github.com/noedaka/go-url-shortener/internal/service"
)

type ExampleMockStorage struct {
	urls  map[string]string
	users map[string]map[string]string
}

func NewExampleMockStorage() *ExampleMockStorage {
	return &ExampleMockStorage{
		urls:  make(map[string]string),
		users: make(map[string]map[string]string),
	}
}

func (m *ExampleMockStorage) Save(ctx context.Context, shortURL, originalURL, userID string) error {
	m.urls[shortURL] = originalURL
	if _, exists := m.users[userID]; !exists {
		m.users[userID] = make(map[string]string)
	}
	m.users[userID][shortURL] = originalURL
	return nil
}

func (m *ExampleMockStorage) Get(ctx context.Context, shortURL string) (string, error) {
	url, exists := m.urls[shortURL]
	if !exists {
		return "", fmt.Errorf("URL not found")
	}
	return url, nil
}

func (m *ExampleMockStorage) GetByUser(ctx context.Context, userID string) ([]model.URLPair, error) {
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

func (m *ExampleMockStorage) Close() error {
	return nil
}

func (m *ExampleMockStorage) DeleteByUser(ctx context.Context, userID string, shortURLs []string) error {
	if userURLs, exists := m.users[userID]; exists {
		for _, shortURL := range shortURLs {
			delete(userURLs, shortURL)
		}
	}
	return nil
}

func (m *ExampleMockStorage) AddURL(shortID, originalURL string) {
	m.urls[shortID] = originalURL
}

func (m *ExampleMockStorage) AddURLForUser(shortID, originalURL, userID string) {
	m.urls[shortID] = originalURL
	if _, exists := m.users[userID]; !exists {
		m.users[userID] = make(map[string]string)
	}
	m.users[userID][shortID] = originalURL
}

// Вспомогательная функция для создания запроса с аутентификацией
func createAuthenticatedRequest(method, url, body string, userID string) *http.Request {
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, url, bytes.NewBufferString(body))
	} else {
		req = httptest.NewRequest(method, url, nil)
	}

	ctx := context.WithValue(req.Context(), config.UserIDKey, userID)
	req = req.WithContext(ctx)

	if body != "" && method == http.MethodPost {
		req.Header.Set("Content-Type", "application/json")
	}

	return req
}

// ExampleHandler_ShortenURLHandler демонстрирует использование эндпоинта для сокращения URL через форму
func ExampleHandler_ShortenURLHandler() {
	mockStorage := NewExampleMockStorage()
	svc := service.NewShortenerService(mockStorage, "http://localhost:8080")
	h := NewHandler(*svc, nil)

	r := chi.NewRouter()
	r.Post("/", h.ShortenURLHandler)

	// Создаем аутентифицированный запрос
	req := createAuthenticatedRequest(http.MethodPost, "/", "https://example.com", "test-user")
	req.Header.Set("Content-Type", "text/plain")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	fmt.Printf("Status: %d\n", rr.Code)
	fmt.Printf("Response contains short URL: %v\n", bytes.Contains(rr.Body.Bytes(), []byte("http://localhost:8080/")))

	// Output:
	// Status: 201
	// Response contains short URL: true
}

// ExampleHandler_APIShortenerHandler демонстрирует использование эндпоинта для сокращения URL
func ExampleHandler_APIShortenerHandler() {
	mockStorage := NewExampleMockStorage()
	svc := service.NewShortenerService(mockStorage, "http://localhost:8080")
	h := NewHandler(*svc, nil)

	r := chi.NewRouter()
	r.Post("/api/shorten", h.APIShortenerHandler)

	requestData := map[string]string{"url": "https://example.com"}
	jsonData, _ := json.Marshal(requestData)

	req := createAuthenticatedRequest(http.MethodPost, "/api/shorten", string(jsonData), "test-user")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	fmt.Printf("Status: %d\n", rr.Code)
	fmt.Printf("Content-Type: %s\n", rr.Header().Get("Content-Type"))
	fmt.Printf("Response contains result: %v\n", bytes.Contains(rr.Body.Bytes(), []byte("result")))

	// Output:
	// Status: 201
	// Content-Type: application/json
	// Response contains result: true
}

// ExampleHandler_ShortIDHandler демонстрирует использование эндпоинта для получения оригинального URL
func ExampleHandler_ShortIDHandler() {
	mockStorage := NewExampleMockStorage()
	// Предварительно добавляем URL в хранилище
	mockStorage.AddURL("abc123", "https://example.com")

	svc := service.NewShortenerService(mockStorage, "http://localhost:8080")
	h := NewHandler(*svc, nil)

	r := chi.NewRouter()
	r.Get("/{id}", h.ShortIDHandler)

	req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	fmt.Printf("Status: %d\n", rr.Code)
	fmt.Printf("Location: %s\n", rr.Header().Get("Location"))

	// Output:
	// Status: 307
	// Location: https://example.com
}

// ExampleHandler_ShortenBatchHandler демонстрирует использование эндпоинта для пакетного сокращения URL
func ExampleHandler_ShortenBatchHandler() {
	mockStorage := NewExampleMockStorage()
	svc := service.NewShortenerService(mockStorage, "http://localhost:8080")
	h := NewHandler(*svc, nil)

	r := chi.NewRouter()
	r.Post("/api/shorten/batch", h.ShortenBatchHandler)

	batchRequest := []model.BatchRequest{
		{CorrelationID: "1", URL: "https://example.com/1"},
		{CorrelationID: "2", URL: "https://example.com/2"},
	}
	jsonData, _ := json.Marshal(batchRequest)

	req := createAuthenticatedRequest(http.MethodPost, "/api/shorten/batch", string(jsonData), "test-user")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	var responses []model.BatchResponse
	json.Unmarshal(rr.Body.Bytes(), &responses)

	fmt.Printf("Status: %d\n", rr.Code)
	fmt.Printf("Content-Type: %s\n", rr.Header().Get("Content-Type"))
	fmt.Printf("Number of responses: %d\n", len(responses))

	// Output:
	// Status: 201
	// Content-Type: application/json
	// Number of responses: 2
}

// ExampleHandler_APIUserUrlsHandler демонстрирует использование эндпоинта для получения URL пользователя
func ExampleHandler_APIUserUrlsHandler() {
	mockStorage := NewExampleMockStorage()
	// Добавляем тестовые данные для пользователя
	mockStorage.AddURLForUser("short1", "https://example.com/1", "test-user")
	mockStorage.AddURLForUser("short2", "https://example.com/2", "test-user")

	svc := service.NewShortenerService(mockStorage, "http://localhost:8080")
	h := NewHandler(*svc, nil)

	r := chi.NewRouter()
	r.Get("/api/user/urls", h.APIUserUrlsHandler)

	req := createAuthenticatedRequest(http.MethodGet, "/api/user/urls", "", "test-user")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	var urls []model.URLPair
	json.Unmarshal(rr.Body.Bytes(), &urls)

	fmt.Printf("Status: %d\n", rr.Code)
	fmt.Printf("Number of user URLs: %d\n", len(urls))

	// Output:
	// Status: 200
	// Number of user URLs: 2
}

// ExampleHandler_APIDeleteShortURLSHandler демонстрирует использование эндпоинта для удаления URL
func ExampleHandler_APIDeleteShortURLSHandler() {
	mockStorage := NewExampleMockStorage()
	svc := service.NewShortenerService(mockStorage, "http://localhost:8080")
	h := NewHandler(*svc, nil)

	r := chi.NewRouter()
	r.Delete("/api/user/urls", h.APIDeleteShortURLSHandler)

	deleteRequest := []string{"short1", "short2"}
	jsonData, _ := json.Marshal(deleteRequest)

	req := createAuthenticatedRequest(http.MethodDelete, "/api/user/urls", string(jsonData), "test-user")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	fmt.Printf("Status: %d\n", rr.Code)

	// Output:
	// Status: 202
}
