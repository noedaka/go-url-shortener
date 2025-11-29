package main

import (
	"fmt"
	"log"
	_ "net/http/pprof"

	"github.com/noedaka/go-url-shortener/internal/app"
)

func main() {
	fmt.Printf("Build version: %s\n", BuildVersion)
	fmt.Printf("Build date: %s\n", BuildDate)
	fmt.Printf("Build commit: %s\n", BuildCommit)

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

/*
  Пример сборки с использованием флагов

  go build -ldflags "\
    -X main.buildVersion=v1.0.0 \
    -X main.buildDate=$(date +'%Y.%m.%d_%H:%M:%S') \
    -X main.buildCommit=$(git rev-parse HEAD)" \
    -o shortener cmd/shortener/main.go
*/
