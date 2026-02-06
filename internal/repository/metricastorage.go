package repository

import (
	"github.com/qesterrx/tplmetrics/internal/model"
)

type MetricaStorage interface {
	//Метод для обновления данных метрики
	UpdateMetrica(model.Metrica) error
	//Метод получения экземпляра метрики по имени
	Metrica(name string) (model.Metrica, error)
	//Получение всех метрик
	AllMetrica() []model.Metrica
	//Отладочный вызов
	ShowDebug()
}
