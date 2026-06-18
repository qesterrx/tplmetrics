// Пакет handler содержит в себе методы для маршрутизации и обработки HTTP запросов сервером.
package handler

import (
	"crypto/rsa"

	"github.com/go-chi/chi"
	"github.com/qesterrx/tplmetrics/internal/middleware"
	"github.com/qesterrx/tplmetrics/internal/service"
)

// HandlerContainer - структура введения зависимостей для хендлеров
// Содержит сслыку на структуру слоя сервиса а так же дополнительные параметры, необходимые обработчикам
type HandlerContainer struct {
	tcl              *service.TCLService
	secretKeyForSign string
	privateKeyPem    *rsa.PrivateKey
}

// NewHandlerContainer - возвращает новый экземпляр HandlerContainer
func NewHandlerContainer(tcl *service.TCLService, secretKeyForSign string, privateKeyRSA *rsa.PrivateKey) *HandlerContainer {

	hc := HandlerContainer{
		tcl:              tcl,
		secretKeyForSign: secretKeyForSign,
		privateKeyPem:    privateKeyRSA,
	}

	return &hc
}

// GetRouter Функция возвращающая роутер запросов который занимается маршрутизацией
// Дополнительно в роутере прописаны исползуемые Middleware функции
// IPContext - функция для добавления в контекст переменной ИП адреса клиента
// LoggingMiddleware - логгирование запросов
// HMACSignMiddleware - расшифровка HMAC сообщения от клиента
// GzipCompressMiddleware - архивирование/разархивирование тела запроса
// RSADecrypt - расшифровка сообщения приватным ключем
func (hc *HandlerContainer) GetRouter() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.IPContext)
	r.Use(middleware.LoggingMiddleware)
	if hc.secretKeyForSign != "" {
		r.Use(middleware.HMACSignMiddleware(hc.secretKeyForSign)) //Это важно! подпись вычислялась после сжатия, значит проверять ее надо ДО распаковки
	}
	if hc.privateKeyPem != nil {
		r.Use(middleware.RSADecrypt(hc.privateKeyPem)) //Шифрование должно быть до архивации
	}
	r.Use(middleware.GzipCompressMiddleware)

	r.Get(`/value/{kind}/{name}`, hc.GetMetricaHandler)
	r.Get(`/`, hc.GetAllCurrentMetricsHandler)
	r.Get(`/ping`, hc.PingDBHandler)
	r.Post(`/update/{kind}/{name}/{value}`, hc.UpdateMetricaHandler)
	r.Post(`/update/`, hc.UpdateMetricaJSONHandler)
	r.Post(`/value/`, hc.GetMetricaJSONHandler)
	r.Post(`/updates/`, hc.UpdateMetricsJSONHandler)

	return r
}
