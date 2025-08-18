package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5/middleware"
)

var compressibleContentTypes = map[string]bool{
	"application/json": true,
	"text/html":        true,
}

type gzipResponseWriter struct {
	middleware.WrapResponseWriter
	gzipWriter *gzip.Writer
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.gzipWriter.Write(b)
}

func (w *gzipResponseWriter) Close() error {
	return w.gzipWriter.Close()
}

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			defer gz.Close()
			r.Body = gz
		}

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		acceptsGzip := strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")
		if acceptsGzip {
			contentType := ww.Header().Get("Content-Type")
			if compressibleContentTypes[contentType] {
				ww.Header().Set("Content-Encoding", "gzip")
				gz := gzip.NewWriter(ww)
				defer gz.Close()

				gzipW := &gzipResponseWriter{
					WrapResponseWriter: ww,
					gzipWriter:         gz,
				}
				next.ServeHTTP(gzipW, r)
				return
			}
		}

		next.ServeHTTP(ww, r)
	})
}
