// Пакет middleware содержит в себе Middleware-фунции сервера.
// Данные функции расширяют функционал без необходимости доработки хендлеров.
package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/qesterrx/tplmetrics/internal/logger"
	"google.golang.org/grpc"
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

// LoggingMiddleware - Middleware-фунция логирования http запросов
// Уровень логирования задан Info
// Отображает:
// URI - какой адрес вызывается
// method - метод обращения
// duration - длительность выполнения
// code - код результата оброботки
// size request - размер запроса (в байтах)
// size response - размер ответа (в байтах)
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

// IPRequestInterceptor - middleware для GRPC
func LoggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()

	// Вызываем основной обработчик
	resp, err := handler(ctx, req)

	// Логируем результат
	duration := time.Since(start)

	result := "OK"
	if err != nil {
		result = err.Error()
	}

	logger.Log.Info().
		Str("RPC", info.FullMethod).
		Str("method", "GRPC").
		Str("duration", duration.String()).
		Str("result", result).
		Send()

	return resp, err
}
