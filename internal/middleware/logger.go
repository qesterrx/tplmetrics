package middleware

import (
	"net/http"
	"time"

	"github.com/qesterrx/tplmetrics/internal/logger"
)

// Middleware
type logResponseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (lrw *logResponseWriter) Write(msg []byte) (int, error) {
	size, err := lrw.ResponseWriter.Write(msg)
	lrw.size = size
	return size, err
}

func (lrw *logResponseWriter) WriteHeader(statusCode int) {
	lrw.ResponseWriter.WriteHeader(statusCode)
	lrw.statusCode = statusCode
}

func LoggingMiddleware(h http.Handler) http.Handler {
	loggedHandler := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		lrw := logResponseWriter{ResponseWriter: w}

		h.ServeHTTP(&lrw, r)

		duration := time.Since(start)

		logger.Log.Info().
			Str("URI", r.RequestURI).
			Str("method", r.Method).
			Str("duration", duration.String()).
			Int("code", lrw.statusCode).
			Int64("size request", r.ContentLength).
			Int("size response", lrw.size).
			Send()

	}

	return http.HandlerFunc(loggedHandler)
}
