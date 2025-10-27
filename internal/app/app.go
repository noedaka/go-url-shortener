package app

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/noedaka/go-url-shortener/internal/audit"
	"github.com/noedaka/go-url-shortener/internal/config"
	dbc "github.com/noedaka/go-url-shortener/internal/config/db"
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

	auditManager := audit.NewAuditManager()
    defer auditManager.Close()

    if cfg.AuditFile != "" {
        fileObserver, err := audit.NewFileObserver(cfg.AuditFile)
        if err != nil {
            logger.Log.Sugar().Errorf("Failed to create file audit observer: %v", err)
        } else {
            auditManager.RegisterObserver(fileObserver)
            logger.Log.Sugar().Infof("File audit enabled: %s", cfg.AuditFile)
        }
    }

    if cfg.AuditURL != "" {
        httpObserver := audit.NewHTTPObserver(cfg.AuditURL)
        auditManager.RegisterObserver(httpObserver)
        logger.Log.Sugar().Infof("HTTP audit enabled: %s", cfg.AuditURL)
    }


	var db *sql.DB
	var store storage.URLStorage

	if isDB {
		var err error
		db, err = sql.Open("pgx", cfg.DatabaseDSN)
		if err != nil {
			return err
		}
		defer db.Close()

		if err := dbc.InitDatabase(db); err != nil {
			return err
		}

		store, err = storage.NewPostgresStorage(db)
		if err != nil {
			return err
		}

		logger.Log.Sugar().Infof("Base Database DSN is %s", cfg.DatabaseDSN)
	} else {
		store = storage.NewFileStorage(cfg.FileStoragePath)
		logger.Log.Sugar().Infof("Base file storage is %s", cfg.FileStoragePath)
	}

	service := service.NewShortenerService(store, cfg.BaseURL)
	handlerURL := handler.NewHandler(*service, db)

	r.Route("/", func(r chi.Router) {
		r.Use(middleware.LoggingMiddleware)
		r.Use(middleware.GzipMiddleware)
		r.Use(middleware.AuthMiddleware)
		r.Use(middleware.AuditMiddleware(auditManager))
		r.Route("/api", func(r chi.Router) {
			r.Route("/shorten", func(r chi.Router) {
				r.Post("/", handlerURL.APIShortenerHandler)
				r.Post("/batch", handlerURL.ShortenBatchHandler)
			})
			r.Route("/user/urls", func(r chi.Router) {
				r.Get("/", handlerURL.APIUserUrlsHandler)
				r.Delete("/", handlerURL.APIDeleteShortURLSHandler)
			})
		})
		r.Post("/", handlerURL.ShortenURLHandler)
		r.Get("/{id}", handlerURL.ShortIDHandler)
		r.Get("/ping", handlerURL.PingDBHandler)
	})

	if err := http.ListenAndServe(cfg.ServerAddress, r); err != nil {
		return err
	}

	return nil
}
