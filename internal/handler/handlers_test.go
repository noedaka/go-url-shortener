package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/noedaka/go-url-shortener/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_ShortenURLHandler(t *testing.T) {
	storage := service.NewURLStorage()
	h := NewHandler(storage, "http://localhost:8080")

	r := chi.NewRouter()
	r.Post("/*", h.ShortenURLHandler)

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
				contains:   "http://",
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
			req := httptest.NewRequest(tt.method, "/"+tt.body, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "text/plain")

			req.Host = "localhost:8080"
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			require.Equal(t, tt.want.statusCode, rr.Code, "Wrong status code: Got %d, expected %d",
				rr.Code, tt.want.statusCode)

			if tt.want.contains != "" {
				assert.Contains(t, rr.Body.String(), tt.want.contains, "Wrong answer: Got %s, expected %s",
					rr.Body.String(), tt.want.contains)
			}

		})
	}
}

func TestHandler_ShortIdHandler(t *testing.T) {
	storage := service.NewURLStorage()
	h := NewHandler(storage, "http://localhost:8080")

	r := chi.NewRouter()
	r.Get("/{id}", h.ShortIDHandler)

	type want struct {
		statusCode int
		location   string
	}

	tests := []struct {
		name   string
		method string
		path   string
		want   want
	}{
		{
			name:   "GET valid",
			method: http.MethodGet,
			path:   "/",
			want: want{
				statusCode: http.StatusTemporaryRedirect,
				location:   "https://example.com",
			},
		},
		{
			name:   "GET nonexistent id",
			method: http.MethodGet,
			path:   "/nonexistentid",
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:   "Wrong method",
			method: http.MethodPut,
			path:   "/nonexistentid",
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request

			if tt.want.location != "" {
				shortID, err := storage.ShortenURL(tt.want.location)
				require.NoError(t, err, "Failed to shorten URL")
				req = httptest.NewRequest(tt.method, "/"+shortID, nil)
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}

			req.Host = "localhost:8080"
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			require.Equal(t, tt.want.statusCode, rr.Code, "Wrong status code: Got %d, expected %d",
				rr.Code, tt.want.statusCode)

			if tt.want.location != "" {
				recLocation := rr.Header().Get("Location")
				assert.Equal(t, tt.want.location, recLocation, "Wrong location header:"+
					"Got %s,  expected %s", tt.want.location, recLocation)
			}

		})
	}
}
