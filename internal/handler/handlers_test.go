package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/noedaka/go-url-shortener/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_ShortenHandler(t *testing.T) {
	storage := service.NewURLStorage()
	h := NewHandler(storage)

	type want struct {
		statusCode int
		contains   string
		location   string
	}

	tests := []struct {
		name   string
		method string
		path   string
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
				statusCode: http.StatusBadRequest,
			},
		},
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
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			switch tt.method {
			case http.MethodPost:
				req = httptest.NewRequest(tt.method, "/", bytes.NewBufferString(tt.body))
				req.Header.Set("Content-Type", "text/plain")
			case http.MethodGet:
				if tt.want.location != "" {
					shortID, err := storage.ShortenURL(tt.want.location)
					require.NoError(t, err, "Failed to shorten URL")
					req = httptest.NewRequest(tt.method, "/"+shortID, nil)
				} else {
					req = httptest.NewRequest(tt.method, tt.path, nil)
				}
			default:
				req = httptest.NewRequest(tt.method, "/", nil)
			}

			req.Host = "localhost:8080"
			rr := httptest.NewRecorder()

			h.ShortenHandler(rr, req)

			require.Equal(t, tt.want.statusCode, rr.Code, "Wrong status code: Got %d, expected %d",
				rr.Code, tt.want.statusCode)

			if tt.want.contains != "" {
				assert.Contains(t, rr.Body.String(), tt.want.contains, "Wrong answer: Got %s, expected %s",
					rr.Body.String(), tt.want.contains)
			}

			if tt.want.location != "" {
				recLocation := rr.Header().Get("Location")
				assert.Equal(t, tt.want.location, recLocation, "Wrong location header:"+
					"Got %s,  expected %s", tt.want.location, recLocation)
			}

		})
	}
}
