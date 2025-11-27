package app

import (
	"database/sql"
	"net/http"
	"net/http/pprof"
	"path/filepath"
	"runtime"

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
	"go.uber.org/zap"
)

func Run() error {
	r := chi.NewRouter()

	if err := logger.Init(); err != nil {
		return err
	}
	logger.Log.Sync()

	cfg, err := config.Init()
	if err != nil {
		return err
	}

	if err := cfg.ValidateConfig(); err != nil {
		return err
	}

	logger.Log.Info("server started",
		zap.String("address", cfg.ServerAddress),
		zap.String("base_url", cfg.BaseURL))

	auditManager := audit.NewAuditManager()
	defer auditManager.Close()

	if cfg.AuditFile != "" {
		fileObserver, err := audit.NewFileObserver(cfg.AuditFile)
		if err != nil {
			logger.Log.Error("failed to create file audit observer",
				zap.Error(err),
				zap.String("file", cfg.AuditFile))
		} else {
			auditManager.RegisterObserver(fileObserver)
			logger.Log.Info("file audit enabled",
				zap.String("file address", cfg.AuditFile))
		}
	}

	if cfg.AuditURL != "" {
		httpObserver := audit.NewHTTPObserver(cfg.AuditURL)
		auditManager.RegisterObserver(httpObserver)
		logger.Log.Info("HTTP audit enabled",
			zap.String("HTTP address", cfg.AuditURL))
	}

	var db *sql.DB
	var store storage.URLStorage

	if cfg.HasDatabase {
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

		logger.Log.Info("config inited",
			zap.String("database dsn", cfg.DatabaseDSN))
	} else {
		store = storage.NewFileStorage(cfg.FileStoragePath)
		logger.Log.Info("config inited",
			zap.String("file storage", cfg.FileStoragePath))
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

	// pprof routing
	r.Route("/debug/pprof", func(r chi.Router) {
		r.Get("/", pprof.Index)
		r.Get("/cmdline", pprof.Cmdline)
		r.Get("/profile", pprof.Profile)
		r.Get("/symbol", pprof.Symbol)
		r.Get("/trace", pprof.Trace)
		r.Get("/heap", pprof.Handler("heap").ServeHTTP)
		r.Get("/goroutine", pprof.Handler("goroutine").ServeHTTP)
		r.Get("/block", pprof.Handler("block").ServeHTTP)
		r.Get("/threadcreate", pprof.Handler("threadcreate").ServeHTTP)
		r.Get("/allocs", pprof.Handler("allocs").ServeHTTP)
	})

	if cfg.EnableHTTPS {
		certFile, keyFile := getCertPaths()
		err = http.ListenAndServeTLS(
			cfg.ServerAddress,
			certFile,
			keyFile,
			nil,
		)
		if err != nil {
			return err
		}

	} else {
		if err := http.ListenAndServe(cfg.ServerAddress, r); err != nil {
			return err
		}
	}

	return nil
}

func getCertPaths() (certFile, keyFile string) {
	_, currentFile, _, _ := runtime.Caller(0)

	projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(currentFile)))

	certsDir := filepath.Join(projectRoot, "cmd", "tls", "certs")
	certFile = filepath.Join(certsDir, "cert.pem")
	keyFile = filepath.Join(certsDir, "private.pem")

	return
}
