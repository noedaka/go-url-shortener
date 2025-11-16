package config

import (
	"flag"
	"fmt"
	"net"
	"net/url"

	"github.com/caarlos0/env/v6"
	"github.com/noedaka/go-url-shortener/internal/model"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	AuditFile       string `env:"AUDIT_FILE"`
	AuditURL        string `env:"AUDIT_URL"`
	HasDatabase     bool
}

const UserIDKey model.ContextKey = "user_id"

const (
	defaultServerAddress   = "localhost:8080"
	defaultBaseURL         = "http://localhost:8080"
	defaultFileStoragePath = "urls.json"
)

func Init() (*Config, error) {
	cfg := &Config{}

	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}

	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "HTTP server adress")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "Base URL")
	flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "File storage path")
	flag.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "Database DSN")
	flag.StringVar(&cfg.AuditFile, "audit-file", cfg.AuditFile, "Audit file")
	flag.StringVar(&cfg.AuditURL, "audit-url", cfg.AuditURL, "Audit URL")
	flag.Parse()

	if cfg.ServerAddress == "" {
		cfg.ServerAddress = defaultServerAddress
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}

	if cfg.DatabaseDSN != "" {
		cfg.HasDatabase = true
		return cfg, nil
	}

	if cfg.FileStoragePath == "" {
		cfg.FileStoragePath = defaultFileStoragePath
	}

	return cfg, nil
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
