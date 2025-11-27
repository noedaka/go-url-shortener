package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"
	"reflect"

	"github.com/caarlos0/env/v6"
	"github.com/noedaka/go-url-shortener/internal/model"
)

const UserIDKey model.ContextKey = "user_id"

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" json:"server_address"`
	BaseURL         string `env:"BASE_URL" json:"base_url"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" json:"file_storage_path"`
	DatabaseDSN     string `env:"DATABASE_DSN" json:"database_dsn"`
	AuditFile       string `env:"AUDIT_FILE" json:"audit_file"`
	AuditURL        string `env:"AUDIT_URL" json:"audit_url"`
	EnableHTTPS     bool   `env:"ENABLE_HTTPS" json:"enable_https"`
	ConfigFile      string `env:"CONFIG"`

	HasDatabase bool
}

func Init() (*Config, error) {
	cfg := &Config{}

	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}

	cfg.bindFlags()
	flag.Parse()

	if cfg.ConfigFile != "" {
		configFile, err := cfg.readConfigFile()
		if err != nil {
			return nil, err
		}

		cfg.mergeConfigs(configFile)
	}

	if cfg.DatabaseDSN != "" {
		cfg.HasDatabase = true
		return cfg, nil
	}

	cfg.setDefaults()

	return cfg, nil
}

func (cfg *Config) setDefaults() {
	if cfg.ServerAddress == "" {
		cfg.ServerAddress = "localhost:8080"
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "http://localhost:8080"
	}

	if cfg.FileStoragePath == "" {
		cfg.FileStoragePath = "urls.json"
	}
}

func (cfg *Config) bindFlags() {
	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "HTTP server adress")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "Base URL")
	flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "File storage path")
	flag.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "Database DSN")
	flag.StringVar(&cfg.AuditFile, "audit-file", cfg.AuditFile, "Audit file")
	flag.StringVar(&cfg.AuditURL, "audit-url", cfg.AuditURL, "Audit URL")
	flag.BoolVar(&cfg.EnableHTTPS, "s", cfg.EnableHTTPS, "Enable HTTPS")
	flag.StringVar(&cfg.ConfigFile, "c", cfg.ConfigFile, "Config file path")
}

func (cfg *Config) readConfigFile() (*Config, error) {
	if _, err := os.Stat(cfg.ConfigFile); os.IsNotExist(err) {
		return nil, nil
	}

	file, err := os.ReadFile(cfg.ConfigFile)
	if err != nil {
		return nil, err
	}

	if len(file) == 0 {
		return nil, nil
	}

	var fileConfig *Config
	if err := json.Unmarshal(file, &fileConfig); err != nil {
		return nil, err
	}

	return fileConfig, nil
}

func (cfg *Config) mergeConfigs(override *Config) {
	valDefault := reflect.ValueOf(cfg).Elem()
	valOverride := reflect.ValueOf(override).Elem()

	for i := 0; i < valDefault.NumField(); i++ {
		fieldDefault := valDefault.Field(i)
		fieldOverride := valOverride.Field(i)

		if fieldDefault.IsZero() && !fieldOverride.IsZero() {
			fieldDefault.Set(fieldOverride)
		}
	}
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
