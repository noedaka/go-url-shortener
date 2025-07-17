package handler

import (
	"io"
	"net/http"
	"strings"

	"github.com/noedaka/go-url-shortener/internal/service"
)

type Handler struct {
	service *service.URLStorage
}

func NewHandler(service *service.URLStorage) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ShortenHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		defer r.Body.Close()

		originalURL := string(body)

		if originalURL == "" {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		shortID, err := h.service.ShortenURL(originalURL)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		shortURL := "http://" + r.Host + "/" + shortID

		w.Header().Set("Content-type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shortURL))
	case http.MethodGet:
		shortID := strings.TrimPrefix(r.URL.Path, "/")

		URL, err := h.service.GetURL(shortID)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		w.Header().Set("Location", URL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
