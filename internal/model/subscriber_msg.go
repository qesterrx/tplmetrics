package model

// UpdateMetricaSubscriberMsg - Структура для лога отправляемая агенту а так же сохраняемая в файле
type UpdateMetricaSubscriberMsg struct {
	TS      int64    `json:"ts"`                   // unix timestamp события
	Metrics []string `json:"metrics"`              // наименование полученных метрик
	IP      string   `json:"ip_address,omitempty"` // IP адрес входящего запроса
}
