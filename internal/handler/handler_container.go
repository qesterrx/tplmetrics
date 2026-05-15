package handler

import (
	"github.com/go-chi/chi"
	"github.com/qesterrx/tplmetrics/internal/middleware"
	"github.com/qesterrx/tplmetrics/internal/service"
)

type HandlerContainer struct {
	tcl              *service.TCLService
	secretKeyForSign string
}

func NewHandlerContainer(tcl *service.TCLService, secretKeyForSign string) *HandlerContainer {
	return &HandlerContainer{tcl: tcl, secretKeyForSign: secretKeyForSign}
}

func (hc *HandlerContainer) GetRouter() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.LoggingMiddleware)
	if hc.secretKeyForSign != "" {
		r.Use(middleware.HMACSignMiddleware(hc.secretKeyForSign)) //Это важно! подпись вычислялась после сжатия, значит проверять ее надо ДО распаковки
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
