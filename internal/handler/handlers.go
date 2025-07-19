package handler

import (
	"io"
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

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading body", http.StatusBadRequest)
	}

	defer r.Body.Close()

	originalURL := string(body)

	shortID, err := h.service.ShortenURL(originalURL)
	if err != nil {
		http.Error(w, "Error shortening url", http.StatusBadRequest)
		return
	}

	shortURL := "http://" + r.Host + "/" + shortID

	w.Header().Set("Content-type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

func (h *Handler) ShortIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	shortID := chi.URLParam(r, "id")

	URL, err := h.service.GetURL(shortID)
	if err != nil {
		http.Error(w, "Error getting url from id", http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", URL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
