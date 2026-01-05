// Модуль handler предоставляет хэндлеры для сервиса сокращения URL.
package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/noedaka/go-url-shortener/internal/config"
	"github.com/noedaka/go-url-shortener/internal/middleware"
	"github.com/noedaka/go-url-shortener/internal/model"
	"github.com/noedaka/go-url-shortener/internal/service"
)

// Handler предоставляет методы для обработки HTTP-запросов.
type Handler struct {
	service service.ShortenerService
	db      *sql.DB
}

// NewHandler создает новый экземпляр Handler.
func NewHandler(service service.ShortenerService, db *sql.DB) *Handler {
	return &Handler{service: service, db: db}
}

// ShortenURLHandler создает короткий URL из переданного URL.
//
// Принимает text/plain, возвращает короткий URL в text/plain.
//
// POST /
func (h *Handler) ShortenURLHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "cannot read body", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	originalURL := string(body)
	userID, ok := getUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	shortID, err := h.service.ShortenURL(r.Context(), originalURL, userID)
	if err != nil {
		if h.handleShortenError(w, err, "text/plain") {
			return
		}
		http.Error(w, "cannot shorten URL", http.StatusInternalServerError)
		return
	}

	middleware.LogAuditEvent(r.Context(), "shorten", originalURL)

	shortURL := h.service.BaseURL + "/" + shortID

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

// APIShortenerHandler создает короткий URL из переданного URL.
//
// Принимает application/json, возвращает короткий URL в application/json.
//
// POST /api/shorten
func (h *Handler) APIShortenerHandler(w http.ResponseWriter, r *http.Request) {
	var req model.Request

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "cannot decode request JSON body", http.StatusBadRequest)
		return
	}

	userID, ok := getUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	shortID, err := h.service.ShortenURL(r.Context(), req.URL, userID)
	if err != nil {
		if h.handleShortenError(w, err, "application/json") {
			return
		}
		http.Error(w, "cannot shorten url", http.StatusInternalServerError)
		return
	}

	middleware.LogAuditEvent(r.Context(), "shorten", req.URL)

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

// APIUserUrlsHandler возвращает все сокращенные текущим пользователем пары URL.
//
// Возвращает короткие и оригинальные URL в application/json.
//
// GET /user/urls
func (h *Handler) APIUserUrlsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	urlPairs, err := h.service.GetURLByUser(r.Context(), userID)
	if err != nil {
		http.Error(w, "cannot get urls by user", http.StatusInternalServerError)
		return
	}

	if len(urlPairs) == 0 {
		http.Error(w, "user did not shorten any urls", http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	enc := json.NewEncoder(w)
	if err := enc.Encode(urlPairs); err != nil {
		http.Error(w, "error encoding response", http.StatusInternalServerError)
		return
	}
}

// ShortIDHandler делает редирект на оригинальный URL по его короткой версии.
//
// GET /{id}
func (h *Handler) ShortIDHandler(w http.ResponseWriter, r *http.Request) {
	shortID := chi.URLParam(r, "id")

	URL, err := h.service.GetURL(r.Context(), shortID)
	if err != nil {
		http.Error(w, "cannot get url from id", http.StatusBadRequest)
		return
	}

	if URL == "" {
		w.WriteHeader(http.StatusGone)
		return
	}

	middleware.LogAuditEvent(r.Context(), "follow", URL)

	w.Header().Set("Location", URL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

// ShortenBatchHandler создает короткие URL каждому переданному URL.
//
// Принимает application/json/ возвращает batchResponse в application/json.
//
// POST /api/shortem/batch
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

	userID, ok := getUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	batchResponse, err := h.service.ShortenMultipleURLS(r.Context(), batchRequest, userID)
	if err != nil {
		http.Error(w, "cannot shorten multiple urls", http.StatusInternalServerError)
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

// APIDeleteShortURLSHandler удаляет все переданные сокращенные URL текущего пользователя.
//
// Принимает application/json
//
// DELETE /user/urls
func (h *Handler) APIDeleteShortURLSHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var shortURLS []string

	if err := json.NewDecoder(r.Body).Decode(&shortURLS); err != nil {
		http.Error(w, "cannot decode request JSON body", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteShortURLSByUser(r.Context(), userID, shortURLS); err != nil {
		http.Error(w, "cannot shorten url", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusAccepted)
}

// PingDBHandler пингует БД.
func (h *Handler) PingDBHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := h.db.PingContext(ctx); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) StatsHandler(trustedSubnet string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ipStr := r.Header.Get("X-Real-IP")
		ip := net.ParseIP(ipStr)

		if ip == nil || !isIPInSubnet(ip, trustedSubnet) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		stats, err := h.service.GetStats(r.Context())
		if err != nil {
			http.Error(w, "Error getting stats", http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		enc := json.NewEncoder(w)
		if err := enc.Encode(stats); err != nil {
			http.Error(w, "error encoding response", http.StatusInternalServerError)
			return
		}
	}
}

func isIPInSubnet(ip net.IP, subnet string) bool {
	if checkIP := net.ParseIP(subnet); checkIP != nil {
		return ip.Equal(checkIP)
	}

	_, network, err := net.ParseCIDR(subnet)
	if err != nil {
		return false
	}

	return network.Contains(ip)
}

func (h *Handler) handleShortenError(w http.ResponseWriter, err error, contentType string) (handled bool) {
	var uniqueErr *model.UniqueViolationError
	if errors.As(err, &uniqueErr) {
		shortURL := h.service.BaseURL + "/" + uniqueErr.ShortID

		if contentType == "application/json" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			resp := model.Response{Result: shortURL}
			json.NewEncoder(w).Encode(resp)
		} else {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(shortURL))
		}
		return true
	}

	return false
}

func getUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(config.UserIDKey).(string)
	return userID, ok
}
