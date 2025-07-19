package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/noedaka/go-url-shortener/internal/handler"
	"github.com/noedaka/go-url-shortener/internal/service"
)

func main() {
	r := chi.NewRouter()

	service := service.NewURLStorage()
	handlerURL := handler.NewHandler(service)

	r.Route("/", func(r chi.Router) {
		r.Post("/*", handlerURL.ShortenURLHandler)
		r.Get("/{id}", handlerURL.ShortIDHandler)
	})

	err := http.ListenAndServe(`:8080`, r)
	if err != nil {
		panic(err)
	}
}
