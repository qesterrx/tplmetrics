package repository

import (
	"github.com/qesterrx/tplmetrics/internal/model"
)

// Интерфейс используемый в хендлерах
type MetricaStorage interface {
	//Метод для обновления данных метрики
	UpdateMetrica(model.Metrica) error
	//Метод получения экземпляра метрики по имени
	Metrica(name string, kind string) (model.Metrica, error)
	//Получение всех метрик
	AllMetrics() []model.Metrica
	//Отладочный вызов
	Debug()
}
