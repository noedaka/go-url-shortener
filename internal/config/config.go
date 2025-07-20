package config

import "flag"

type Config struct {
	ServerAdress string
	BaseURL      string
}

func Init() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.ServerAdress, "a", "localhost:8080", "HTTP server adress")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost8080", "Base URL")

	return cfg
}
