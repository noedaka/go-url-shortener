package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/noedaka/go-url-shortener/internal/service"
)

type Handler struct {
	service *service.URLStorage
}

func NewHandler(service *service.URLStorage) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ShortenURLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	originalURL := chi.URLParam(r, "*")

	shortID, err := h.service.ShortenURL(originalURL)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	shortURL := "http://" + r.Host + "/" + shortID

	w.Header().Set("Content-type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

func (h *Handler) ShortIdHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	shortID := chi.URLParam(r, "id")

	URL, err := h.service.GetURL(shortID)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", URL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
