package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/noedaka/go-url-shortener/internal/config"
	"github.com/noedaka/go-url-shortener/internal/handler"
	"github.com/noedaka/go-url-shortener/internal/logger"
	"github.com/noedaka/go-url-shortener/internal/middleware"
	"github.com/noedaka/go-url-shortener/internal/service"
)

func Run() error {
	r := chi.NewRouter()

	if err := logger.Init(); err != nil {
		return err
	}
	logger.Log.Sync()

	cfg := config.Init()
	if err := cfg.ValidateConfig(); err != nil {
		return err
	}

	logger.Log.Sugar().Infof("Server is on %s", cfg.ServerAddress)
	logger.Log.Sugar().Infof("Base URL is %s", cfg.BaseURL)

	service := service.NewURLStorage()
	handlerURL := handler.NewHandler(service, cfg.BaseURL)

	r.Route("/", func(r chi.Router) {
		r.Use(middleware.LoggingMiddleware)
		r.Post("/api/shortener", handlerURL.APIShortenerHandler)
		r.Post("/*", handlerURL.ShortenURLHandler)
		r.Get("/{id}", handlerURL.ShortIDHandler)
	})

	if err := http.ListenAndServe(cfg.ServerAddress, r); err != nil {
		return err
	}

	return nil
}
