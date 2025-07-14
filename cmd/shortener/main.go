package main

import (
	"net/http"

	"github.com/noedaka/go-url-shortener/internal/handler"
	"github.com/noedaka/go-url-shortener/internal/service"
)

func main() {
	mux := http.NewServeMux()

	service := service.NewURLStorage()
	handlerURL := handler.NewHandler(service)

	mux.HandleFunc(`/`, handlerURL.ShortenHandler)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
