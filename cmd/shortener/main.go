package main

import (
	"log"

	"github.com/noedaka/go-url-shortener/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
