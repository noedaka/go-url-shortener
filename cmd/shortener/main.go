package main

import (
	"fmt"
	"log"
	_ "net/http/pprof"

	"github.com/noedaka/go-url-shortener/internal/app"
)

func main() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

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
