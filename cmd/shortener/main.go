package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/noedaka/go-url-shortener/internal/config"
	"github.com/noedaka/go-url-shortener/internal/handler"
	"github.com/noedaka/go-url-shortener/internal/service"
)

func main() {
	r := chi.NewRouter()

	cfg := config.Init()
	flag.Parse()

	err := cfg.ValidateConfig()

	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	log.Printf("Server is on %s", cfg.ServerAddress)
	log.Printf("Base URL is %s", cfg.BaseURL)

	service := service.NewURLStorage()
	handlerURL := handler.NewHandler(service, cfg.BaseURL)

	r.Route("/", func(r chi.Router) {
		r.Post("/*", handlerURL.ShortenURLHandler)
		r.Get("/{id}", handlerURL.ShortIDHandler)
	})

	err = http.ListenAndServe(cfg.ServerAddress, r)
	if err != nil {
		log.Fatalf("Fatal server error: %v", err)
	}
}
