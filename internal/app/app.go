package app

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/noedaka/go-url-shortener/internal/config"
	"github.com/noedaka/go-url-shortener/internal/handler"
	"github.com/noedaka/go-url-shortener/internal/logger"
	"github.com/noedaka/go-url-shortener/internal/middleware"
	"github.com/noedaka/go-url-shortener/internal/service"
	"github.com/noedaka/go-url-shortener/internal/storage"
)

func Run() error {
	r := chi.NewRouter()

	if err := logger.Init(); err != nil {
		return err
	}
	logger.Log.Sync()

	cfg, isDB := config.Init()
	if err := cfg.ValidateConfig(); err != nil {
		return err
	}

	logger.Log.Sugar().Infof("Server is on %s", cfg.ServerAddress)
	logger.Log.Sugar().Infof("Base URL is %s", cfg.BaseURL)

	db, err := sql.Open("pgx", cfg.DatabaseDSN)
	if err != nil {
		return err
	}
	defer db.Close()

	var store storage.URLStorage

	if isDB {
		store, err = storage.NewPostgresStorage(db)
		if err != nil {
			return err
		}

		logger.Log.Sugar().Infof("Base Database DSN is %s", cfg.DatabaseDSN)
	} else {
		store = storage.NewFileStorage(cfg.FileStoragePath)
		logger.Log.Sugar().Infof("Base file storage is %s", cfg.FileStoragePath)
	}

	service := service.NewShortenerService(store)
	handlerURL := handler.NewHandler(*service, cfg.BaseURL, db)

	r.Route("/", func(r chi.Router) {
		r.Use(middleware.LoggingMiddleware)
		r.Use(middleware.GzipMiddleware)
		r.Post("/api/shorten", handlerURL.APIShortenerHandler)
		r.Post("/", handlerURL.ShortenURLHandler)
		r.Get("/{id}", handlerURL.ShortIDHandler)
		r.Get("/ping", handlerURL.PingDBHandler)
	})

	if err := http.ListenAndServe(cfg.ServerAddress, r); err != nil {
		return err
	}

	return nil
}
