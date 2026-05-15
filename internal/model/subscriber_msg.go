package model

// Структура для общения между агентом и сервером а так же для хранения данных в файле
type UpdateMetricaSubscriberMsg struct {
	TS      int64    `json:"ts"`                   // unix timestamp события
	Metrics []string `json:"metrics"`              // наименование полученных метрик
	IP      string   `json:"ip_address,omitempty"` // IP адрес входящего запроса
}
