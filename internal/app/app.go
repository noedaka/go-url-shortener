package app

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

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

			r.Route("/internal", func(r chi.Router) {
				r.Use(func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						if cfg.TrustedSubnet == "" {
							http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
							return
						}
						next.ServeHTTP(w, r)
					})
				})
				r.Get("/stats", handlerURL.StatsHandler(cfg.TrustedSubnet))
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv := http.Server{Addr: cfg.ServerAddress, Handler: r}

	serverErr := make(chan error, 1)
	go func() {
		if cfg.EnableHTTPS {
			certFile, keyFile := getCertPaths()
			serverErr <- srv.ListenAndServeTLS(certFile, keyFile)
		} else {
			serverErr <- srv.ListenAndServe()
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	select {
	case err := <-serverErr:
		// Сервер сам завершился с ошибкой
		return err
	case sig := <-sigChan:
		// Получен сигнал на завершение
		logger.Log.Info("received signal, starting graceful shutdown",
			zap.String("signal", sig.String()))

		shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 30*time.Second)
		defer shutdownCancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Log.Error("server shutdown error", zap.Error(err))
			// Принудительное закрытие сервера
			if err := srv.Close(); err != nil {
				logger.Log.Error("server force close error", zap.Error(err))
			}

			return err
		}

		logger.Log.Info("server stopped gracefully")
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
