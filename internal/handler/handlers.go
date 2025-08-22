package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/noedaka/go-url-shortener/internal/model"
	"github.com/noedaka/go-url-shortener/internal/service"
)

type Handler struct {
	service service.ShortenerService
	db      *sql.DB
}

func NewHandler(service service.ShortenerService, db *sql.DB) *Handler {
	return &Handler{service: service, db: db}
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

	shortURL := h.service.BaseURL + "/" + shortID

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

	shortURL := h.service.BaseURL + "/" + shortID

	resp := model.Response{
		Result: shortURL,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	enc := json.NewEncoder(w)
	if err := enc.Encode(resp); err != nil {
		http.Error(w, "error encoding response", http.StatusInternalServerError)
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

func (h *Handler) PingDBHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := h.db.PingContext(ctx); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) ShortenBatchHandler(w http.ResponseWriter, r *http.Request) {
	var batchRequest []model.BatchRequest

	if err := json.NewDecoder(r.Body).Decode(&batchRequest); err != nil {
		http.Error(w, "cannot decode request JSON body", http.StatusBadRequest)
		return
	}

	if len(batchRequest) == 0 {
		http.Error(w, "empty JSON body", http.StatusBadRequest)
		return
	}

	batchResponse, err := h.service.ShortenMultipleURLS(batchRequest)
	if err != nil {
		http.Error(w, "cannot shorten multiple urls", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	enc := json.NewEncoder(w)
	if err := enc.Encode(batchResponse); err != nil {
		http.Error(w, "error encoding response", http.StatusInternalServerError)
		return
	}
}
