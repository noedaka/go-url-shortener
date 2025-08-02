package config

import (
	"flag"
	"fmt"
	"net"
	"net/url"
)

type Config struct {
	ServerAddress string
	BaseURL       string
}

func Init() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "HTTP server adress")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "Base URL")

	return cfg
}

func (cfg *Config) ValidateConfig() error {
	if cfg.ServerAddress == "" {
		return fmt.Errorf("server address cannot be empty")
	}

	_, _, err := net.SplitHostPort(cfg.ServerAddress)
	if err != nil {
		return fmt.Errorf("invalid server address format: %w", err)
	}

	if cfg.BaseURL == "" {
		return fmt.Errorf("base URL cannot be empty")
	}

	u, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return fmt.Errorf("invalid base URL: %s", u)
	}

	return nil
}
