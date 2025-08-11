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
}

const (
	defaultServerAddress   = "localhost:8080"
	defaultBaseURL         = "http://localhost:8080"
	defaultFileStoragePath = "urls.json"
)

func Init() *Config {
	cfg := &Config{}

	err := env.Parse(cfg)
	if err != nil {
		log.Fatal(err)
	}

	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "HTTP server adress")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "Base URL")
	flag.StringVar(&cfg.BaseURL, "f", cfg.BaseURL, "File strage path")
	flag.Parse()

	if cfg.ServerAddress == "" {
		cfg.ServerAddress = defaultServerAddress
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}
	if cfg.FileStoragePath == "" {
		cfg.FileStoragePath = defaultFileStoragePath
	}

	return cfg
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
