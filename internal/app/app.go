package app

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/noedaka/go-url-shortener/internal/config"
	"github.com/noedaka/go-url-shortener/internal/handler"
	"github.com/noedaka/go-url-shortener/internal/service"
)

func Run() error {
	r := chi.NewRouter()

	cfg := config.Init()

	if err := cfg.ValidateConfig(); err != nil {
		return err
	}

	log.Printf("Server is on %s", cfg.ServerAddress)
	log.Printf("Base URL is %s", cfg.BaseURL)

	service := service.NewURLStorage()
	handlerURL := handler.NewHandler(service, cfg.BaseURL)

	r.Route("/", func(r chi.Router) {
		r.Post("/*", handlerURL.ShortenURLHandler)
		r.Get("/{id}", handlerURL.ShortIDHandler)
	})

	if err := http.ListenAndServe(cfg.ServerAddress, r); err != nil {
		return err
	}

	return nil
}
