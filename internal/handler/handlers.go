package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/noedaka/go-url-shortener/internal/model"
	"github.com/noedaka/go-url-shortener/internal/service"
)

type Handler struct {
	service service.URLStorer
	baseURL string
}

func NewHandler(service service.URLStorer, baseURL string) *Handler {
	return &Handler{service: service, baseURL: baseURL}
}

func (h *Handler) ShortenURLHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "cannot read body", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	originalURL := string(body)

	shortID, err := h.service.ShortenURL(originalURL)
	if err != nil {
		http.Error(w, "cannot shorten url", http.StatusBadRequest)
		return
	}

	shortURL := h.baseURL + "/" + shortID

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

func (h *Handler) APIShortenerHandler(w http.ResponseWriter, r *http.Request) {
	var req model.Request

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "cannot decode request JSON body", http.StatusBadRequest)
		return
	}

	shortID, err := h.service.ShortenURL(req.URL)
	if err != nil {
		http.Error(w, "cannot shorten url", http.StatusBadRequest)
		return
	}

	shortURL := h.baseURL + "/" + shortID

	resp := model.Response{
		Result: shortURL,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	enc := json.NewEncoder(w)
	if err := enc.Encode(resp); err != nil {
		http.Error(w, "error encoding response", http.StatusBadRequest)
		return
	}
}

func (h *Handler) ShortIDHandler(w http.ResponseWriter, r *http.Request) {
	shortID := chi.URLParam(r, "id")

	URL, err := h.service.GetURL(shortID)
	if err != nil {
		http.Error(w, "cannot get url from id", http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", URL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
