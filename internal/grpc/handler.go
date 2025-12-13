package grpc

import (
	"context"
	"fmt"

	"github.com/noedaka/go-url-shortener/api/proto"
	"github.com/noedaka/go-url-shortener/internal/config"
	"github.com/noedaka/go-url-shortener/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Handler обрабатывает gRPC запросы
type handler struct {
	proto.UnimplementedShortenerServiceServer
	service service.ShortenerService
	baseURL string
}

// NewHandler создает новый gRPC хендлер
func newHandler(service service.ShortenerService, baseURL string) *handler {
	return &handler{
		service: service,
		baseURL: baseURL,
	}
}

// ShortenURL обрабатывает запрос на сокращение URL
func (h *handler) ShortenURL(ctx context.Context, req *proto.URLShortenRequest) (*proto.URLShortenResponse, error) {
	userID, ok := getUserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	shortID, err := h.service.ShortenURL(ctx, req.GetUrl(), userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot shorten URL: %v", err)
	}

	shortURL := fmt.Sprintf("%s/%s", h.baseURL, shortID)

	var response proto.URLShortenResponse
	response.SetResult(shortURL)

	return &response, nil
}

// ExpandURL обрабатывает запрос на получение оригинального URL
func (h *handler) ExpandURL(ctx context.Context, req *proto.URLExpandRequest) (*proto.URLExpandResponse, error) {
	originalURL, err := h.service.GetURL(ctx, req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot get URL: %v", err)
	}

	if originalURL == "" {
		return nil, status.Error(codes.NotFound, "URL has been deleted")
	}

	var response proto.URLExpandResponse
	response.SetResult(originalURL)

	return &response, nil
}

// ListUserURLs обрабатывает запрос на получение всех URL пользователя
func (h *handler) ListUserURLs(ctx context.Context, req *emptypb.Empty) (*proto.UserURLsResponse, error) {
	userID, ok := getUserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	urlPairs, err := h.service.GetURLByUser(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot get URLs by user: %v", err)
	}

	URLs := make([]*proto.URLData, 0, len(urlPairs))
	for _, pair := range urlPairs {
		shortURL := fmt.Sprintf("%s/%s", h.baseURL, pair.ShortURL)

		var URL proto.URLData
		URL.SetShortUrl(shortURL)
		URL.SetOriginalUrl(pair.OriginalURL)

		URLs = append(URLs, &URL)
	}

	var response proto.UserURLsResponse
	response.SetUrl(URLs)

	return &response, nil
}

func getUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(config.UserIDKey).(string)
	return userID, ok
}
