package repository

import (
	"context"
	"time"

	"github.com/qesterrx/tplmetrics/internal/logger"
	"github.com/qesterrx/tplmetrics/internal/model"
)

type MetricaStorageMode string

const (
	MetricaStorageModeSync  MetricaStorageMode = "sync"
	MetricaStorageModeAsync MetricaStorageMode = "async"
)

// Интерфейс используемый в хендлерах
type MetricaStorage interface {
	//Метод для обновления данных метрики
	UpdateMetrica(model.Metrica) error
	//Обновление массива метрик
	UpdateMetricaBatch([]model.Metrica) error
	//Метод получения экземпляра метрики по имени
	Metrica(name string, kind string) (model.Metrica, error)
	//Получение всех метрик
	AllMetrics() []model.Metrica
	//Метод для "сброса" накопившихся записей в долговременное хранилище
	WriteMetrics() error
	//Проверка хранилища
	Check() error
	//Отладочный вызов
	Debug()
}

// Переодическая запись текущего состояния в хранилище
func TickerWriteMetrics(ctx context.Context, storage MetricaStorage, interval int) {

	ticker := time.NewTicker(time.Second * time.Duration(interval))
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Log.Debug().Msg("Завершение TickerWriteMetrics по контексту")
			//Перед тем как выйти сделаем запись
			err := storage.WriteMetrics()
			if err != nil {
				logger.Log.Error().Msg("TickerWriteMetrics Ошибка при сохранении данных в файл" + err.Error())
			}
			return
		case <-ticker.C:
			err := storage.WriteMetrics()
			if err != nil {
				logger.Log.Error().Msg("TickerWriteMetrics Ошибка при сохранении данных в файл" + err.Error())
			}
			logger.Log.Debug().Msg("TickerWriteMetrics сброс данных в хранилище")
		}
	}
}
