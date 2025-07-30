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
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
}

const (
	defaultServerAddress = "localhost:8080"
	defaultBaseURL       = "http://localhost:8080"
)

func Init() *Config {
	cfg := &Config{}

	err := env.Parse(cfg)
	if err != nil {
		log.Fatal(err)
	}

	if cfg.ServerAddress == "" {
		flag.StringVar(&cfg.ServerAddress, "a", defaultServerAddress, "HTTP server adress")
	}

	if cfg.BaseURL == "" {
		flag.StringVar(&cfg.BaseURL, "b", defaultBaseURL, "Base URL")
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
