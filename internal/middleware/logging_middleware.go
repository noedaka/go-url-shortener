package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/noedaka/go-url-shortener/internal/logger"
	"go.uber.org/zap"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		logger.Log.Info("request",
			zap.String("path", r.RequestURI),
			zap.String("method", r.Method),
			zap.Duration("duration", time.Since(start)),
		)

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		next.ServeHTTP(ww, r)

		logger.Log.Info("answer",
			zap.Int("status code", ww.Status()),
			zap.Int("bytes", ww.BytesWritten()),
		)
	})
}
