package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/noedaka/go-url-shortener/cmd/reset/generator"
)

func main() {
	rootDir := "./../.."
	if len(os.Args) > 1 {
		rootDir = os.Args[1]
	}

	absRoot, err := filepath.Abs(rootDir)
	if err != nil {
		log.Fatalf("Error getting absolute path: %v\n", err)
	}

	generator, err := generator.NewGenerator()
	if err != nil {
		log.Fatalf("Error creating generator: %v\n", err)
	}

	if err := generator.WalkAndProcess(absRoot); err != nil {
		log.Fatalf("Error during generation: %v\n", err)
	}
}
