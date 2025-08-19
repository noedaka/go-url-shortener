package config

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/url"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
}

const (
	defaultServerAddress   = "localhost:8080"
	defaultBaseURL         = "http://localhost:8080"
	defaultFileStoragePath = "urls.json"
)

func Init() (*Config, bool) {
	cfg := &Config{}

	err := env.Parse(cfg)
	if err != nil {
		log.Fatal(err)
	}

	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "HTTP server adress")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "Base URL")
	flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "File storage path")
	flag.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "Database DSN")
	flag.Parse()

	cfg.DatabaseDSN = "postgres://postgres:admin@localhost:5432/url_shortener?sslmode=disable"

	if cfg.ServerAddress == "" {
		cfg.ServerAddress = defaultServerAddress
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}

	if cfg.DatabaseDSN != "" {
		return cfg, true
	}

	if cfg.FileStoragePath == "" {
		cfg.FileStoragePath = defaultFileStoragePath
	}

	return cfg, false
}

func (cfg *Config) ValidateConfig() error {
	_, _, err := net.SplitHostPort(cfg.ServerAddress)
	if err != nil {
		return fmt.Errorf("invalid server address format: %w", err)
	}

	u, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return fmt.Errorf("invalid base URL: %s", u)
	}

	return nil
}
